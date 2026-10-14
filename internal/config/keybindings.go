package config

import (
	"fmt"
	"os"
	"strings"
)

type Action string

const (
	ActionMoveUp       Action = "move_up"
	ActionMoveDown     Action = "move_down"
	ActionMoveLeft     Action = "move_left"
	ActionMoveRight    Action = "move_right"
	ActionJumpFirstRow Action = "jump_first_row"
	ActionJumpLastRow  Action = "jump_last_row"
	ActionJumpFirstCol Action = "jump_first_col"
	ActionJumpLastCol  Action = "jump_last_col"
	ActionPageUp       Action = "page_up"
	ActionPageDown     Action = "page_down"

	ActionVisualMode     Action = "visual_mode"
	ActionVisualLineMode Action = "visual_line_mode"

	ActionYank      Action = "yank"
	ActionExport    Action = "export"
	ActionExportAll Action = "export_all"
	ActionEnter     Action = "enter"
	ActionUpdate    Action = "update"
	ActionDeleteRow Action = "delete_row"

	ActionEditSQL      Action = "edit_sql"
	ActionSaveQuery    Action = "save_query"
	ActionSearch       Action = "search"
	ActionSearchCol    Action = "search_col"
	ActionNextMatch    Action = "next_match"
	ActionPrevMatch    Action = "prev_match"
	ActionPrevColMatch Action = "prev_col_match"
	ActionNextColMatch Action = "next_col_match"

	ActionHelp         Action = "help"
	ActionToggleFooter Action = "toggle_footer"
	ActionQuit         Action = "quit"

	ActionDetailClose      Action = "detail_close"
	ActionDetailEdit       Action = "detail_edit"
	ActionDetailOpen       Action = "detail_open"
	ActionDetailYank       Action = "detail_yank"
	ActionDetailScrollUp   Action = "detail_scroll_up"
	ActionDetailScrollDown Action = "detail_scroll_down"

	ActionExplorerSelect     Action = "explorer_select"
	ActionExplorerColumns    Action = "explorer_columns"
	ActionExplorerRelations  Action = "explorer_relations"
	ActionExplorerToggleFold Action = "explorer_toggle_fold"

	ActionHelpClose Action = "help_close"
)

type Mode int

const (
	ModeNormal Mode = iota
	ModeDetail
	ModeHelp
	ModeExplorer
	ModeExplorerRelations
)

var modeActions = map[Mode][]Action{
	ModeNormal: {
		ActionMoveUp, ActionMoveDown, ActionMoveLeft, ActionMoveRight,
		ActionJumpFirstRow, ActionJumpLastRow, ActionJumpFirstCol, ActionJumpLastCol,
		ActionPageUp, ActionPageDown,
		ActionVisualMode, ActionVisualLineMode,
		ActionYank, ActionExport, ActionExportAll, ActionEnter,
		ActionUpdate, ActionDeleteRow,
		ActionEditSQL, ActionSaveQuery, ActionSearch, ActionSearchCol,
		ActionNextMatch, ActionPrevMatch, ActionPrevColMatch, ActionNextColMatch,
		ActionHelp, ActionToggleFooter, ActionQuit,
	},
	ModeDetail: {
		ActionDetailClose, ActionDetailEdit, ActionDetailOpen, ActionDetailYank,
		ActionDetailScrollUp, ActionDetailScrollDown,
	},
	ModeHelp: {
		ActionHelpClose,
	},
	ModeExplorer: {
		ActionMoveUp, ActionMoveDown,
		ActionJumpFirstRow, ActionJumpLastRow,
		ActionPageUp, ActionPageDown,
		ActionSearch, ActionNextMatch, ActionPrevMatch,
		ActionExplorerSelect,
		ActionExplorerColumns, ActionExplorerRelations, ActionExplorerToggleFold,
		ActionHelp, ActionToggleFooter, ActionQuit,
	},
	ModeExplorerRelations: {
		ActionMoveUp, ActionMoveDown,
		ActionQuit,
	},
}

var DefaultKeybindings = map[Action][]string{
	ActionMoveUp:       {"k", "up"},
	ActionMoveDown:     {"j", "down"},
	ActionMoveLeft:     {"h", "left"},
	ActionMoveRight:    {"l", "right"},
	ActionJumpFirstRow: {"g"},
	ActionJumpLastRow:  {"G"},
	ActionJumpFirstCol: {"0", "home", "_"},
	ActionJumpLastCol:  {"$", "end"},
	ActionPageUp:       {"ctrl+u", "pgup"},
	ActionPageDown:     {"ctrl+d", "pgdown"},

	ActionVisualMode:     {"v"},
	ActionVisualLineMode: {"V"},

	ActionYank:      {"y"},
	ActionExport:    {"x"},
	ActionExportAll: {"X"},
	ActionEnter:     {"enter"},
	ActionUpdate:    {"u"},
	ActionDeleteRow: {"D"},

	ActionEditSQL:      {"e"},
	ActionSaveQuery:    {"s"},
	ActionSearch:       {"/"},
	ActionSearchCol:    {"f"},
	ActionNextMatch:    {"n"},
	ActionPrevMatch:    {"N"},
	ActionPrevColMatch: {","},
	ActionNextColMatch: {";"},

	ActionHelp:         {"H"},
	ActionToggleFooter: {"?"},
	ActionQuit:         {"q", "ctrl+c"},

	ActionDetailClose:      {"q", "esc", "enter"},
	ActionDetailEdit:       {"u"},
	ActionDetailOpen:       {"o"},
	ActionDetailYank:       {"y"},
	ActionDetailScrollUp:   {"k", "up"},
	ActionDetailScrollDown: {"j", "down"},

	ActionExplorerSelect:     {"enter"},
	ActionExplorerColumns:    {"c"},
	ActionExplorerRelations:  {"r"},
	ActionExplorerToggleFold: {"tab"},

	ActionHelpClose: {"H", "q", "esc"},
}

type KeyMap struct {
	ActionToKeys map[Action][]string
	keyByMode    map[Mode]map[string]Action
}

func BuildKeyMap(userOverrides map[string][]string) *KeyMap {
	actionToKeys := make(map[Action][]string)
	for action, keys := range DefaultKeybindings {
		copied := make([]string, len(keys))
		copy(copied, keys)
		actionToKeys[action] = copied
	}

	for actionName, keys := range userOverrides {
		action := Action(actionName)
		if _, exists := DefaultKeybindings[action]; !exists {
			fmt.Fprintf(os.Stderr, "warning: unknown keybinding '%s' in config\n", actionName)
			continue
		}
		filtered := make([]string, 0, len(keys))
		for _, k := range keys {
			if k == "" {
				fmt.Fprintf(os.Stderr, "warning: empty key binding for '%s' ignored\n", actionName)
				continue
			}
			filtered = append(filtered, k)
		}
		if len(filtered) > 0 {
			actionToKeys[action] = filtered
		}
	}

	keyByMode := make(map[Mode]map[string]Action)
	for mode, actions := range modeActions {
		km := make(map[string]Action)
		for _, action := range actions {
			for _, key := range actionToKeys[action] {
				if existing, exists := km[key]; exists && existing != action {
					fmt.Fprintf(os.Stderr, "warning: key '%s' bound to both '%s' and '%s' in same mode\n",
						key, existing, action)
				}
				km[key] = action
			}
		}
		keyByMode[mode] = km
	}

	return &KeyMap{
		ActionToKeys: actionToKeys,
		keyByMode:    keyByMode,
	}
}

func (km *KeyMap) ResolveKey(mode Mode, key string) (Action, bool) {
	action, ok := km.keyByMode[mode][key]
	return action, ok
}

func (km *KeyMap) DisplayKeys(action Action) string {
	keys := km.ActionToKeys[action]
	return strings.Join(keys, " / ")
}

func (km *KeyMap) FirstKey(action Action) string {
	keys := km.ActionToKeys[action]
	if len(keys) > 0 {
		return keys[0]
	}
	return string(action)
}
