package commands

import (
	"fmt"
	"io"
	"log/slog"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type styles struct {
	title        lipgloss.Style
	item         lipgloss.Style
	selectedItem lipgloss.Style
	pagination   lipgloss.Style
	help         lipgloss.Style
	quitText     lipgloss.Style
}

func newStyles(darkBG bool, width int) styles {
	var s styles
	s.title = lipgloss.NewStyle().MarginLeft(2)
	s.item = lipgloss.NewStyle().PaddingLeft(4).Width(width)
	s.selectedItem = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170")).Width(width)
	s.pagination = list.DefaultStyles(darkBG).PaginationStyle.PaddingLeft(4)
	s.help = list.DefaultStyles(darkBG).HelpStyle.PaddingLeft(4).PaddingBottom(1)
	s.quitText = lipgloss.NewStyle().Margin(1, 0, 2, 4)
	return s
}

type itemDelegate struct {
	styles *styles
}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	str := fmt.Sprintf("%d. %s - %s", index+1, listItem.FilterValue(), "")
	switch v := listItem.(type) {
	case ModelItem:
		str = fmt.Sprintf("%s/%s", v.ProviderID, v.ModelID)
	case CommandItem:
		str = fmt.Sprintf("%s - %s", v.Name, v.Desc)
	}

	fn := d.styles.item.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return d.styles.selectedItem.Render(("> " + strings.Join(s, " ")))
		}
	}

	if _, err := fmt.Fprint(w, fn(str)); err != nil {
		slog.Error("render_item", "error", err)
	}
}
