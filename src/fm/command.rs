use crate::log;
use std::process::Command as cmd;

use super::{state::State, tree::Tree};

#[derive(Default)]
pub struct Command {
    pub string: String,
}

impl Command {
    pub fn reset(&mut self) {
        self.string = String::new();
    }
    pub fn shell(&mut self) -> Result<(), String> {
        let shell = std::env::var("SHELL").map_err(|e| format!("could not get editor {e}"))?;

        let mut child = cmd::new(shell)
            .spawn()
            .map_err(|e| format!("failed to start editor {e}"))?;

        child.wait().map_err(|e| format!("child failed {e}"))?;
        Ok(())
    }

    pub fn edit(&mut self, tree: &mut Tree, state: &mut State) -> Result<(), String> {
        // TODO: better error handling
        let editor = std::env::var("EDITOR").map_err(|e| format!("could not get editor {e}"))?;

        let file = tree
            .cwd()
            .get_selected_file(state.show_hidden, &state.query_string)
            .unwrap();
        if !file.is_dir {
            let mut child = cmd::new(editor)
                .arg(file.name)
                .spawn()
                .map_err(|e| format!("failed to start editor {e}"))?;

            child.wait().map_err(|e| format!("child failed {e}"))?;
        }
        Ok(())
    }
    pub fn execute(&mut self, tree: &mut Tree, state: &mut State) {
        // | :delete       | ed      | Moves file to temporary location. After fm is closed, the files will be deleted permanently
        // | :undo         | eu      | Put files back where they were and don't delete them at the end of the fm session.
        // | :yank         | yy      | Copy file under cursor
        // | :cut          | dd      | Cut file under cursor
        // | :paste        | pp      | Paste file to current directory

        let cmds = self.string.split(' ').collect::<Vec<&str>>();
        match cmds[0] {
            "rename" | "rn" => {
                // TODO: if no cmd[1] ask for name
                if let Some(new_name) = cmds.get(1) {
                    tree.cwd().rename_selected(new_name, state.show_hidden);
                }
            }
            "cd" => {
                log::write(format!("cd {}", cmds[1]));
                let dir = shellexpand::tilde(cmds[1]);
                let dir = std::fs::canonicalize(dir.into_owned()).expect("idk");
                if let Err(e) = std::env::set_current_dir(&dir) {
                    log::error(format!("unknown command {e}"));
                }
                tree.cd(dir);
            }
            "exit" | "q" => {
                log::write("quitting".to_string());
                state.exit();
            }
            "edit" | "e" => {
                log::write("editing".to_string());
                self.edit(tree, state).expect("could not edit");
            }
            "hidden" | "h" => {
                log::write("toggle hidden".to_string());
                state.toggle_hidden();
            }
            _ => match cmds[0].chars().nth(0).unwrap() {
                '!' => {
                    let mut exec = cmds.get(0).unwrap().to_string();
                    exec.remove(0);
                    let args = cmds.get(1..).unwrap();
                    log::write(format!("execute {} {:?}", exec, args));

                    match cmd::new(exec).args(args).spawn() {
                        Ok(mut child) => {
                            if let Err(e) = child.wait() {
                                log::error(e.to_string())
                            }
                        }
                        Err(e) => log::error(e.to_string()),
                    }

                    if let Err(e) = tree.cwd().refresh() {
                        log::error(e);
                    }
                }
                _ => {
                    log::error(format!("unknown command {:?}", cmds));
                }
            },
        }
        self.reset()
    }
}
