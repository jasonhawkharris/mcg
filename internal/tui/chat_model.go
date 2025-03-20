package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/hawk/mcgraph/internal/llm"
)

// Message represents a single message in the chat
type Message struct {
	Content       string  // Full content of the message
	VisibleContent string  // For AI messages, this grows during animation
	IsUser        bool
	Time          time.Time
	IsComplete    bool    // Whether the typing animation is complete
	IsSystem      bool    // Whether this is a system message (not from user or AI)
}

// typingMsg is a message for typing animation ticks
type typingMsg struct{}

// thinkingTickMsg is a message for the "Thinking..." animation
type thinkingTickMsg struct{}

// ChatModel is the main model for the chat TUI
type ChatModel struct {
	messages         []Message
	viewport         viewport.Model
	textarea         textarea.Model
	spinner          spinner.Model
	waitingForResp   bool
	typingActive     bool    // Whether typing animation is active
	typingSpeed      int     // Characters per tick
	thinkingDots     int     // Number of dots in "Thinking..." (1-3)
	err              error
	width            int
	height           int
	ready            bool
	quitting         bool
	db               DBInterface
	conversationID   uuid.UUID
}

// Message styles
var (
	userStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5DADE2")).
			PaddingLeft(2).
			PaddingRight(2).
			Bold(true)

	aiStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#58D68D")).
		Bold(true)

	timestampStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A0A0A0")).
			Italic(true).
			MarginRight(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E74C3C")).
			MarginLeft(2).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F4D03F")).
			MarginLeft(2).
			Italic(true)
			
	systemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF9933")).
			MarginLeft(2).
			Italic(true)
)

// NewChatModel creates a new chat model
func NewChatModel(db DBInterface, conversationID uuid.UUID, loadedMessages []Message) ChatModel {
	// Create a textarea for input
	ta := textarea.New()
	ta.Placeholder = "Ask a question..."
	ta.Focus()
	ta.CharLimit = 10000
	ta.SetWidth(80)
	ta.SetHeight(3)
	ta.ShowLineNumbers = false

	// Create a spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	// Create a viewport for displaying messages
	vp := viewport.New(80, 20)
	vp.SetContent("")

	messages := []Message{}
	
	// Flag to track if we're continuing a conversation
	isContinuingConversation := len(loadedMessages) > 0

	// If we have loaded messages, use those
	if isContinuingConversation {
		messages = loadedMessages
	} else {
		// Add welcome message
		currentLLM := llm.GetCurrentLLM()
		welcomeMsg := fmt.Sprintf("Welcome to McGraph Chat! Current LLM: %s\nType your questions and press Enter to submit.\nType Ctrl+C to quit.", currentLLM)

		messages = []Message{
			{
				Content:       welcomeMsg,
				VisibleContent: welcomeMsg,  // Welcome message appears immediately (no animation)
				IsUser:        false,
				Time:          time.Now(),
				IsComplete:    true,
			},
		}
	}

	model := ChatModel{
		messages:       messages,
		textarea:       ta,
		viewport:       vp,
		spinner:        s,
		waitingForResp: false,
		typingActive:   false,
		typingSpeed:    4,  // Characters per typing tick
		thinkingDots:   1,  // Start with one dot
		db:             db,
		conversationID: conversationID,
	}
	
	// If we're continuing a conversation, we need to update the viewport content
	// and scroll to the bottom once the model is initialized
	if isContinuingConversation {
		// We can't scroll here because viewport size isn't set yet
		// Setting a flag to scroll on first WindowSizeMsg
		model.ready = false
	}
	
	return model
}

// typingAnimation returns a command that sends typing tick messages
func typingAnimation() tea.Cmd {
	return tea.Tick(time.Millisecond*20, func(t time.Time) tea.Msg {
		return typingMsg{}
	})
}

// thinkingAnimation returns a command that animates the "Thinking..." dots
func thinkingAnimation() tea.Cmd {
	return tea.Tick(time.Millisecond*200, func(t time.Time) tea.Msg {
		return thinkingTickMsg{}
	})
}

// Init initializes the model
func (m ChatModel) Init() tea.Cmd {
	// Start with 1 dot
	m.thinkingDots = 1
	return tea.Batch(textarea.Blink, m.spinner.Tick, thinkingAnimation())
}

// Update handles events and updates the model
func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
		spCmd tea.Cmd
	)

	// Handle different message types
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit
		
		case tea.KeyEnter:
			if !m.waitingForResp {
				// Only send if there's content
				input := strings.TrimSpace(m.textarea.Value())
				if input != "" {
					// Check for special commands
					if input == "/summarize" {
						// Don't add the command to the visible messages
						m.textarea.Reset()
						
						// Set waiting state
						m.waitingForResp = true
						
						// Update viewport with a system message
						m.messages = append(m.messages, Message{
							Content:       "Generating conversation summary...",
							VisibleContent: "Generating conversation summary...",
							IsUser:        false,
							Time:          time.Now(),
							IsComplete:    true,
							IsSystem:      true, // Mark as system message
						})
						m.updateViewportContent()
						m.viewport.GotoBottom()
						
						// Generate summary
						return m, m.getSummary()
					}
					
					// Normal message flow
					// Add user message to the UI
					m.messages = append(m.messages, Message{
						Content:       input,
						VisibleContent: input, // User messages show immediately
						IsUser:        true,
						Time:          time.Now(),
						IsComplete:    true,
					})
					
					// Save user message to database
					if m.db != nil {
						ctx := context.Background()
						_, err := m.db.AddMessage(ctx, m.conversationID, "user", input)
						if err != nil {
							// Just log the error, don't interrupt the user experience
							m.err = fmt.Errorf("failed to save message: %w", err)
						}
						
						// Generate title from first message if this is the first message
						if len(m.messages) == 2 { // Welcome message + first user message
							go func() {
								_, err := m.db.GenerateTitle(ctx, m.conversationID)
								if err != nil {
									// Just log the error
									m.err = fmt.Errorf("failed to generate title: %w", err)
								}
							}()
						}
					}
					
					// Clear input
					m.textarea.Reset()
					
					// Set waiting state
					m.waitingForResp = true
					
					// Update viewport with the new message
					m.updateViewportContent()
					
					// Request answer from LLM
					return m, m.getResponse(input)
				}
			}
		}
		
	// Window size changed
	case tea.WindowSizeMsg:
		headerHeight := 1
		footerHeight := 4 // textarea + padding
		
		isFirstResize := !m.ready
		
		if isFirstResize {
			m.width = msg.Width
			m.height = msg.Height
			m.ready = true
		} else {
			m.width = msg.Width
			m.height = msg.Height
		}
		
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - headerHeight - footerHeight
		m.textarea.SetWidth(msg.Width)
		
		m.updateViewportContent()
		
		// If this is the first time we're resizing and we have messages
		// (like when continuing a conversation), scroll to the bottom
		if isFirstResize && len(m.messages) > 0 {
			m.viewport.GotoBottom()
		}
	
	// Response received
	case llmResponse:
		m.waitingForResp = false
		
		if msg.err != nil {
			m.err = msg.err
			errorMessage := fmt.Sprintf("Error: %v", msg.err)
			m.messages = append(m.messages, Message{
				Content:       errorMessage,
				VisibleContent: errorMessage, // Error messages show immediately
				IsUser:        false,
				Time:          time.Now(),
				IsComplete:    true,
			})
			m.updateViewportContent()
			m.viewport.GotoBottom()
		} else {
			// For system responses like summaries, don't save to DB
			if msg.isSystemResponse {
				// Replace the "generating" message with the actual response
				lastMsgIdx := len(m.messages) - 1
				if lastMsgIdx >= 0 && m.messages[lastMsgIdx].IsSystem {
					// Replace the last message
					m.messages[lastMsgIdx] = Message{
						Content:       msg.response,
						VisibleContent: "", // Start empty for typing effect
						IsUser:        false,
						IsSystem:      true,
						Time:          time.Now(),
						IsComplete:    false,
					}
				} else {
					// Add a new message
					m.messages = append(m.messages, Message{
						Content:       msg.response,
						VisibleContent: "", // Start empty for typing effect
						IsUser:        false,
						IsSystem:      true,
						Time:          time.Now(),
						IsComplete:    false,
					})
				}
			} else {
				// Add the message with no visible content initially
				m.messages = append(m.messages, Message{
					Content:       msg.response,
					VisibleContent: "", // Start empty for typing effect
					IsUser:        false,
					Time:          time.Now(),
					IsComplete:    false,
				})
				
				// Save assistant message to database
				if m.db != nil {
					ctx := context.Background()
					_, err := m.db.AddMessage(ctx, m.conversationID, "assistant", msg.response)
					if err != nil {
						// Just log the error, don't interrupt the user experience
						m.err = fmt.Errorf("failed to save message: %w", err)
					}
				}
			}
			
			// Update the viewport to show the empty message
			m.updateViewportContent()
			m.viewport.GotoBottom()
			
			// Start the typing animation
			m.typingActive = true
			return m, typingAnimation()
		}
	
	// Handle spinner ticks while waiting
	case spinner.TickMsg:
		if m.waitingForResp {
			m.spinner, spCmd = m.spinner.Update(msg)
		}
		
	// Handle thinking animation ticks
	case thinkingTickMsg:
		// Always send next tick, we need this animation for "Thinking..." text
		cmd := thinkingAnimation()
		
		// Only update dots when waiting for a response
		if m.waitingForResp {
			// Cycle through 1, 2, 3 dots
			m.thinkingDots = (m.thinkingDots % 3) + 1
		}
		
		return m, cmd

	// Handle typing animation
	case typingMsg:
		if m.typingActive {
			// Find the last message (which should be an AI message being typed)
			lastIdx := len(m.messages) - 1
			if lastIdx >= 0 && !m.messages[lastIdx].IsUser && !m.messages[lastIdx].IsComplete {
				// Get current visible content and full content
				fullContent := m.messages[lastIdx].Content
				currentVisible := m.messages[lastIdx].VisibleContent
				
				// Calculate how many new characters to add
				remainingChars := len(fullContent) - len(currentVisible)
				charsToAdd := m.typingSpeed
				if remainingChars < charsToAdd {
					charsToAdd = remainingChars
				}
				
				if charsToAdd > 0 {
					// Add more characters to visible content
					m.messages[lastIdx].VisibleContent = fullContent[:len(currentVisible)+charsToAdd]
					m.updateViewportContent()
					m.viewport.GotoBottom()
					
					// Continue animation if not complete
					if len(m.messages[lastIdx].VisibleContent) < len(fullContent) {
						return m, typingAnimation()
					}
				}
				
				// Animation complete
				if len(m.messages[lastIdx].VisibleContent) >= len(fullContent) {
					m.messages[lastIdx].IsComplete = true
					m.typingActive = false
				}
			}
		}
	}

	// Handle updates for text input and viewport
	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	return m, tea.Batch(tiCmd, vpCmd, spCmd)
}

// View renders the UI
func (m ChatModel) View() string {
	if m.quitting {
		return ""
	}
	
	if !m.ready {
		return "Initializing..."
	}
	
	// Render the messages viewport
	viewportContent := m.viewport.View()
	
	// Render the input area
	inputArea := m.textarea.View()
	
	// Show spinner if waiting for response
	if m.waitingForResp {
		// Create the animated "Thinking..." text with the correct number of dots
		thinkingText := fmt.Sprintf("Thinking%s", strings.Repeat(".", m.thinkingDots))
		inputArea = fmt.Sprintf("%s %s", m.spinner.View(), thinkingText)
	}
	
	// Put it all together
	return fmt.Sprintf("%s\n\n%s", viewportContent, inputArea)
}

// getResponse requests a response from the LLM
func (m ChatModel) getResponse(input string) tea.Cmd {
	return func() tea.Msg {
		response, err := llm.GetResponse(input)
		return llmResponse{
			response: response,
			err:      err,
		}
	}
}

// getSummary generates a summary of the conversation
func (m ChatModel) getSummary() tea.Cmd {
	return func() tea.Msg {
		// Build a conversation history string
		var historyBuilder strings.Builder
		
		// Skip the welcome message and the "generating summary" message
		for i, msg := range m.messages {
			// Skip system messages and the last message (which is the "generating summary" message)
			if msg.IsSystem || i == len(m.messages)-1 {
				continue
			}
			
			if msg.IsUser {
				historyBuilder.WriteString("User: ")
			} else {
				historyBuilder.WriteString("Assistant: ")
			}
			historyBuilder.WriteString(msg.Content)
			historyBuilder.WriteString("\n\n")
		}
		
		// Create the prompt for generating a summary
		prompt := fmt.Sprintf(`Below is a conversation between a user and an AI assistant. 
Please provide a concise summary (about 3-5 sentences) of the main topics, questions, and information covered in this conversation.

CONVERSATION:
%s

SUMMARY:`, historyBuilder.String())
		
		// Get response from LLM
		response, err := llm.GetResponse(prompt)
		
		return llmResponse{
			response: "# Conversation Summary\n\n" + response,
			err:      err,
			isSystemResponse: true,
		}
	}
}

// llmResponse is a message containing the LLM's response
type llmResponse struct {
	response        string
	err             error
	isSystemResponse bool
}

// updateViewportContent updates the viewport with formatted messages
func (m *ChatModel) updateViewportContent() {
	var sb strings.Builder
	
	for i, msg := range m.messages {
		// Format timestamp
		timestamp := timestampStyle.Render(msg.Time.Format("15:04:05"))
		
		if msg.IsUser {
			// Format user message
			sb.WriteString(fmt.Sprintf("%s %s: %s\n\n", 
				timestamp, 
				userStyle.Render("You"),
				msg.Content))
		} else if msg.IsSystem {
			// Format system message
			sb.WriteString(fmt.Sprintf("%s %s: %s\n\n", 
				timestamp, 
				systemStyle.Render("System"),
				msg.VisibleContent)) // Use visibleContent for animation
		} else {
			// Format AI message with syntax highlighting for code blocks
			// Use the visibleContent for the typing animation effect
			highlightedContent := Highlight(msg.VisibleContent)
			sb.WriteString(fmt.Sprintf("%s %s: %s\n\n", 
				timestamp, 
				aiStyle.Render("McGraph"),
				highlightedContent))
		}
		
		// Add separator between messages, except for the last one
		if i < len(m.messages)-1 {
			sb.WriteString(strings.Repeat("─", m.width) + "\n\n")
		}
	}
	
	m.viewport.SetContent(sb.String())
}