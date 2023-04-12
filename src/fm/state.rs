use crossterm::event::{KeyCode, KeyEvent, KeyModifiers};

use super::command::Command;

#[derive(Default, Eq, PartialEq)]
pub enum Mode {
    #[default]
    Normal,
    Search,
    Command,
}
#[derive(Default, Clone, Copy)]
pub enum Action {
    #[default]
    Nop,
    ResetSelection,
    Parent,
    Selected,
    SelectFile,
    Up,
    Down,
    PgUp,
    PgDown,
    Edit,
    Shell,
    Execute,
}

#[derive(Default)]
pub struct State {
    pub show_hidden: bool,
    pub hide_parent: bool,
    pub action: Action,
    pub query_string: String,
    pub mode: Mode,
    pub exit: bool,
}

impl State {
    pub fn exit(&mut self) -> &mut State {
        self.exit = true;
        self
    }
    fn with_mode(&mut self, mode: Mode) -> &mut State {
        self.mode = mode;
        self
    }
    fn with_action(&mut self, cmd: Action) -> &mut State {
        self.action = cmd;
        self
    }
    pub fn toggle_hidden(&mut self) -> &mut State {
        self.show_hidden = !self.show_hidden;
        self
    }
    pub fn toggle_show_parent(&mut self) -> &mut State {
        self.hide_parent = !self.hide_parent;
        self
    }
    pub fn reset_query(&mut self) -> &mut State {
        self.query_string = String::new();
        self
    }

    pub fn handle_input(&mut self, key: KeyEvent, cmd: &mut Command) -> &mut State {
        self.action = Action::Nop;
        match self.mode {
            Mode::Normal => self.handle_normal(key),
            Mode::Search => self.handle_search(key),
            Mode::Command => self.handle_command(key, cmd),
        }
    }

    fn handle_normal(&mut self, key: KeyEvent) -> &mut State {
        match key.code {
            // modes
            KeyCode::Esc | KeyCode::Char('q') => self.exit(),
            KeyCode::Enter => self.with_action(Action::Edit),
            KeyCode::Tab => self.with_action(Action::SelectFile),
            KeyCode::Char(':') => self.with_mode(Mode::Command),
            KeyCode::Char('/') => self
                .with_mode(Mode::Search)
                .with_action(Action::ResetSelection),
            KeyCode::Char('H') => self.toggle_hidden(),
            KeyCode::Char('P') => self.toggle_show_parent(),
            KeyCode::Char('S') => self.with_action(Action::Shell),
            // motion
            KeyCode::Left | KeyCode::Char('h') => self.with_action(Action::Parent),
            KeyCode::Down | KeyCode::Char('j') => self.with_action(Action::Down),
            KeyCode::Up | KeyCode::Char('k') => self.with_action(Action::Up),
            // FIXME: handle symlinks
            KeyCode::Right | KeyCode::Char('l') => self.with_action(Action::Selected),
            KeyCode::Char(c) => {
                if key.modifiers == KeyModifiers::CONTROL {
                    match c {
                        'u' => self.with_action(Action::PgUp),
                        'd' => self.with_action(Action::PgDown),
                        _ => self,
                    }
                } else {
                    self
                }
            }
            _ => self,
        }
    }

    fn handle_search(&mut self, key: KeyEvent) -> &mut State {
        match key.code {
            KeyCode::Esc => self.reset_query().with_mode(Mode::Normal),
            KeyCode::Enter => self.with_mode(Mode::Normal).with_action(Action::Selected),
            KeyCode::Backspace => {
                self.query_string.pop();
                self
            }
            KeyCode::Char(c) => {
                if key.modifiers == KeyModifiers::CONTROL {
                    match c {
                        'c' => self.reset_query().with_mode(Mode::Normal),
                        'n' => self.with_action(Action::Up),
                        'p' => self.with_action(Action::Down),
                        _ => self,
                    }
                } else {
                    self.query_string.push(c);
                    self
                }
            }
            _ => self,
        }
    }

    fn handle_command(&mut self, key: KeyEvent, cmd: &mut Command) -> &mut State {
        match key.code {
            KeyCode::Esc => self.with_mode(Mode::Normal),
            KeyCode::Enter => self.with_mode(Mode::Normal).with_action(Action::Execute),
            KeyCode::Backspace => {
                cmd.string.pop();
                self
            }
            KeyCode::Char(c) => {
                if key.modifiers == KeyModifiers::CONTROL {
                    match c {
                        'c' => {
                            cmd.reset();
                            self.with_mode(Mode::Normal)
                        }
                        _ => self,
                    }
                } else {
                    cmd.string.push(c);
                    self
                }
            }
            _ => self,
        }
    }
}
