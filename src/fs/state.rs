use crossterm::event::{KeyCode, KeyEvent, KeyModifiers};

#[derive(PartialEq, Default)]
pub enum Mode {
    #[default]
    NORMAL,
    SEARCH,
    COMMAND,
}
#[derive(PartialEq, Default, Clone, Copy)]
pub enum Command {
    #[default]
    Nop,
    Parent,
    Selected,
    Up,
    Down,
}

#[derive(Default)]
pub struct State {
    pub show_hidden: bool,
    pub command: Command,
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
    pub fn handle_input(&mut self, key: KeyEvent) -> &mut State {
        match self.mode {
            Mode::NORMAL => match key.code {
                // modes
                KeyCode::Esc | KeyCode::Char('q') => return self.exit(),
                KeyCode::Char(':') => {
                    self.with_mode(Mode::COMMAND)
                    // self.cwd().state.select(Some(0))
                }
                KeyCode::Char('/') => {
                    self.with_mode(Mode::SEARCH)
                    // self.cwd().state.select(Some(0))
                }
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
                _ => self,
            },
            Mode::SEARCH => match key.code {
                KeyCode::Char(c) => {
                    if key.modifiers == KeyModifiers::CONTROL {
                        match c {
                            'c' => {
                                self.command_string = String::new();
                                self.with_mode(Mode::NORMAL)
                            }
                            _ => self,
                        }
                    } else {
                        self.command_string.push(c);
                        self
                    }
                }
                // this should cd into directories and open files in EDITOR
                KeyCode::Enter => {
                    // TODO: select file not just exit
                    self.command_string = String::new();
                    self.with_mode(Mode::NORMAL)
                }
                KeyCode::Backspace => {
                    self.command_string.pop();
                    self
                }
                KeyCode::Esc => {
                    self.command_string = String::new();
                    self.with_mode(Mode::NORMAL)
                }
                _ => self,
            },
            Mode::COMMAND => match key.code {
                KeyCode::Esc | KeyCode::Char('q') => return self.exit(),

                KeyCode::Char(c) => {
                    if key.modifiers == KeyModifiers::CONTROL {
                        match c {
                            'c' => {
                                self.command_string = String::new();
                                self.with_mode(Mode::NORMAL)
                            }
                            _ => self,
                        }
                    } else {
                        self.command_string.push(c);
                        self
                    }
                }
                _ => self,
            },
        }
    }
}
