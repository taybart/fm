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
    pub fn exit(&mut self) -> &mut State {
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
    pub fn toggle_hidden(&mut self) -> &mut State {
        self.show_hidden = !self.show_hidden;
        self
    }

    pub fn reset_command(&mut self) -> &mut State {
        self.command_string = String::new();
        self
    }
    pub fn reset_query(&mut self) -> &mut State {
        self.query_string = String::new();
        self
    }
    pub fn handle_input(&mut self, key: KeyEvent) -> &mut State {
        self.command = Command::Nop;
        match self.mode {
            Mode::Normal => match key.code {
                // modes
                KeyCode::Esc | KeyCode::Char('q') => self.exit(),
                KeyCode::Enter => self.with_command(Command::Edit),
                KeyCode::Tab => self.with_command(Command::SelectFile),
                KeyCode::Char(':') => self.with_mode(Mode::Command),
                KeyCode::Char('/') => self
                    .with_mode(Mode::Search)
                    .with_command(Command::ResetSelection),
                KeyCode::Char('H') => self.toggle_hidden(),
                KeyCode::Char('S') => self.with_command(Command::Shell),
                // motion
                KeyCode::Left | KeyCode::Char('h') => self.with_command(Command::Parent),
                KeyCode::Down | KeyCode::Char('j') => self.with_command(Command::Down),
                KeyCode::Up | KeyCode::Char('k') => self.with_command(Command::Up),
                // FIXME: handle symlinks
                KeyCode::Right | KeyCode::Char('l') => self.with_command(Command::Selected),
                KeyCode::Char(c) => {
                    if key.modifiers == KeyModifiers::CONTROL {
                        match c {
                            'u' => self.with_command(Command::PgUp),
                            'd' => self.with_command(Command::PgDown),
                            _ => self,
                        }
                    } else {
                        self
                    }
                }
                _ => self,
            },
            Mode::Search => match key.code {
                KeyCode::Esc => self.reset_query().with_mode(Mode::Normal),
                KeyCode::Enter => self.with_mode(Mode::Normal).with_command(Command::Selected),
                KeyCode::Backspace => {
                    self.query_string.pop();
                    self
                }
                KeyCode::Char(c) => {
                    if key.modifiers == KeyModifiers::CONTROL {
                        match c {
                            'c' => self.reset_query().with_mode(Mode::Normal),
                            'n' => self.with_command(Command::Up),
                            'p' => self.with_command(Command::Down),
                            _ => self,
                        }
                    } else {
                        self.query_string.push(c);
                        self
                    }
                }
                _ => self,
            },
            Mode::Command => match key.code {
                KeyCode::Esc => self.with_mode(Mode::Normal),
                KeyCode::Enter => self.with_mode(Mode::Normal).with_command(Command::Execute),
                KeyCode::Backspace => {
                    self.command_string.pop();
                    self
                }
                KeyCode::Char(c) => {
                    if key.modifiers == KeyModifiers::CONTROL {
                        match c {
                            'c' => self.reset_command().with_mode(Mode::Normal),
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
