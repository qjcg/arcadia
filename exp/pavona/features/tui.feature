Feature: TUI framework — tui package
  The tui package wraps bubbletea with layout and keybinding primitives.

  Scenario: A basic TUI renders and responds to keys
    When I run a TUI app
    Then I see a rendered screen
    And pressing "q" exits

  Scenario: Keybinding registry maps keys to actions
    When I register "ctrl+c" → quit and "enter" → submit
    Then pressing ctrl+c exits
    And pressing enter triggers submit

  Scenario: Help overlay is available
    When I press "?"
    Then a help overlay shows all registered keybindings
