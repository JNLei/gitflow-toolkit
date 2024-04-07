package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	selectorTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#2E2E2E", Dark: "#DDDDDD"}).
				Background(lipgloss.AdaptiveColor{Light: "#19A04B", Dark: "#25A065"}).
				Bold(true).
				Padding(0, 1)

	selectorNormalStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#DDDDDD"}).
				Padding(0, 0, 0, 2)

	selectorSelectedStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(lipgloss.AdaptiveColor{Light: "#9F72FF", Dark: "#AD58B4"}).
				Foreground(lipgloss.AdaptiveColor{Light: "#9A4AFF", Dark: "#EE6FF8"}).
				Bold(true).
				Padding(0, 0, 0, 1)

	selectorPaginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)

	selectorHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#6F6C6C", Dark: "#7A7A7A"})
)

type selectorItem struct {
	ct    string
	title string
}

func (cti selectorItem) FilterValue() string { return cti.title }

type selectorDelegate struct{}

func (d selectorDelegate) Height() int                             { return 1 }
func (d selectorDelegate) Spacing() int                            { return 0 }
func (d selectorDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d selectorDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(selectorItem)
	if !ok {
		return
	}
	str := fmt.Sprintf("%d. %s", index+1, i.title)
	if index == m.Index() {
		_, _ = fmt.Fprintf(w, selectorSelectedStyle.Render(str))
	} else {
		_, _ = fmt.Fprintf(w, selectorNormalStyle.Render(str))
	}

}

type selectorModel struct {
	list   list.Model
	choice string
}

func (m selectorModel) Init() tea.Cmd {
	return nil
}

func (m selectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "enter":
			m.choice = m.list.SelectedItem().(selectorItem).ct
			return m, func() tea.Msg { return done{nextView: INPUTS} }

		default:
			if !m.list.SettingFilter() && (keypress == "q" || keypress == "esc") {
				return m, tea.Quit
			}

			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

	default:
		return m, nil
	}
}

func (m selectorModel) View() string {
	if m.choice != "" {
		return m.choice
	}
	return "\n" + m.list.View()
}

func newSelectorModel() selectorModel {
	prioritized := prioritizeCommitType([]selectorItem{
		{ct: feat, title: featDesc},
		{ct: fix, title: fixDesc},
		{ct: docs, title: docsDesc},
		{ct: style, title: styleDesc},
		{ct: refactor, title: refactorDesc},
		{ct: test, title: testDesc},
		{ct: chore, title: choreDesc},
		{ct: perf, title: perfDesc},
		{ct: hotfix, title: hotfixDesc},
	})
	listItems := []list.Item{}
	for _, commitType := range prioritized {
		listItems = append(listItems, commitType)
	}
	l := list.NewModel(listItems, selectorDelegate{}, 20, 12)

	l.Title = "Select Commit Type"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = selectorTitleStyle
	l.Styles.PaginationStyle = selectorPaginationStyle
	h := help.NewModel()
	h.Styles.ShortDesc = selectorHelpStyle
	h.Styles.ShortSeparator = selectorHelpStyle
	h.Styles.ShortKey = selectorHelpStyle
	l.Help = h

	return selectorModel{list: l}
}

func prioritizeCommitType(items []selectorItem) []selectorItem {
	branch, err := currentBranch()
	if err != nil {
		return items
	}

	currentBranchType := strings.Split(branch, "/")[0]
	for ind, branchType := range items {
		if currentBranchType == branchType.ct {
			return append(append([]selectorItem{branchType}, items[:ind]...), items[ind+1:]...)
		}
	}

	return items
}
