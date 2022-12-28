use crossterm::event::{KeyCode, KeyEvent, KeyModifiers};

#[derive(Default, Eq, PartialEq)]
pub enum Mode {
    #[default]
    Normal,
    Search,
    Command,
}
#[derive(Default, Clone, Copy)]
pub enum Command {
    #[default]
    Nop,
    Parent,
    Selected,
    Up,
    Down,
    Edit,
}

#[derive(Default)]
pub struct State {
    pub show_hidden: bool,
    pub command: Command,
    pub query_string: String,
    pub command_string: String,
    pub mode: Mode,
    pub exit: bool,
}

/* States:
 *      cd parent
 *      cd selected
 *      move up
 *      move down
 */

impl Command {}

impl State {
    fn exit(&mut self) -> &mut State {
        self.exit = true;
        self
    }
    fn with_mode(&mut self, mode: Mode) -> &mut State {
        self.mode = mode;
        self
    }
    fn with_command(&mut self, cmd: Command) -> &mut State {
        self.command = cmd;
        self
    }

    pub fn reset_command(&mut self) {
        self.query_string = String::new();
    }
    pub fn handle_input(&mut self, key: KeyEvent) -> &mut State {
        self.command = Command::Nop;
        match self.mode {
            Mode::Normal => match key.code {
                // modes
                KeyCode::Esc | KeyCode::Char('q') => self.exit(),
                KeyCode::Char(':') => {
                    self.with_mode(Mode::Command)
                    // self.cwd().state.select(Some(0))
                }
                KeyCode::Char('/') => self.with_mode(Mode::Search),
                KeyCode::Char('H') => {
                    self.show_hidden = !self.show_hidden;
                    self
                }
                // motion
                KeyCode::Left | KeyCode::Char('h') => self.with_command(Command::Parent),
                KeyCode::Down | KeyCode::Char('j') => self.with_command(Command::Up),
                KeyCode::Up | KeyCode::Char('k') => self.with_command(Command::Down),
                // TODO: handle symlinks
                KeyCode::Right | KeyCode::Char('l') => self.with_command(Command::Selected),
                KeyCode::Enter => self.with_command(Command::Edit),
                _ => self,
            },
            Mode::Search => match key.code {
                KeyCode::Char(c) => {
                    if key.modifiers == KeyModifiers::CONTROL {
                        match c {
                            'c' => {
                                self.reset_command();
                                self.with_mode(Mode::Normal)
                            }
                            'n' => self.with_command(Command::Up),
                            'p' => self.with_command(Command::Down),
                            _ => self,
                        }
                    } else {
                        self.query_string.push(c);
                        self
                    }
                }
                // this should cd into directories and open files in EDITOR
                KeyCode::Enter => {
                    // TODO: select file not just exit
                    // self.with_mode(Mode::Normal).with_command(Command::Edit)
                    // crate::log::write("select".to_string());
                    self.with_mode(Mode::Normal).with_command(Command::Selected)
                }
                KeyCode::Backspace => {
                    self.query_string.pop();
                    self
                }
                KeyCode::Esc => {
                    self.reset_command();
                    self.with_mode(Mode::Normal)
                }
                _ => self,
            },
            Mode::Command => match key.code {
                KeyCode::Esc | KeyCode::Char('q') => self.exit(),

                KeyCode::Char(c) => {
                    if key.modifiers == KeyModifiers::CONTROL {
                        match c {
                            'c' => {
                                self.query_string = String::new();
                                self.with_mode(Mode::Normal)
                            }
                            _ => self,
                        }
                    } else {
                        self.query_string.push(c);
                        self
                    }
                }
                _ => self,
            },
        }
    }
}
