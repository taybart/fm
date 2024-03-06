use anyhow::{Context, Result};

use super::dir::Dir;
use super::state::State;
use std::{
    collections::HashMap,
    env, fs,
    path::{Path, PathBuf},
};

pub struct Tree {
    fs: HashMap<PathBuf, Dir>,
}

impl Tree {
    pub fn new() -> Result<Tree> {
        let parent = Dir::new(Tree::parent_path()?)?;
        let cwd = Dir::new(Tree::cwd_path()?)?;

        let mut fs: HashMap<PathBuf, Dir> = HashMap::new();
        fs.insert(parent.path.clone(), parent);
        fs.insert(cwd.path.clone(), cwd);

        Ok(Tree { fs })
    }

    pub fn parent_path() -> Result<PathBuf> {
        fs::canonicalize(Path::new("..")).context("could not get parent path")
    }
    pub fn cwd_path() -> Result<PathBuf> {
        fs::canonicalize(Path::new(".")).context("could not get cwd path")
    }

    pub fn parent(&mut self) -> &mut Dir {
        let parent_path = Tree::parent_path().expect("could not canonicalize parent path");
        self.fs
            .entry(parent_path.clone())
            .or_insert_with(|| Dir::new(parent_path.clone()).unwrap());
        self.fs.get_mut(&parent_path).unwrap()
    }
    pub fn cwd(&mut self) -> &mut Dir {
        // FIXME: check if folder exists, if not make it
        let cwd_path = fs::canonicalize(Path::new(".")).expect("could not canonicalize cwd path");
        self.fs.get_mut(&cwd_path).unwrap()
    }

    pub fn cd(&mut self, dir: PathBuf) {
        // cd
        env::set_current_dir(Path::new(&dir)).unwrap();

        let cwd_path = Tree::cwd_path().expect("could not canonicalize parent path");

        let parent_path = Tree::parent_path().expect("could not canonicalize parent path");

        self.fs
            .entry(parent_path.clone())
            .or_insert_with(|| Dir::new(parent_path).unwrap());
        self.fs
            .entry(cwd_path.clone())
            .or_insert_with(|| Dir::new(cwd_path.clone()).unwrap());
    }

    pub fn cd_parent(&mut self, state: &State) {
        let init_parent_path = Tree::cwd_path().expect("could not canonicalize parent path");

        env::set_current_dir(Path::new("..")).unwrap();
        let parent_path = Tree::parent_path().expect("could not canonicalize parent path");

        self.fs
            .entry(parent_path.clone())
            .or_insert_with(|| Dir::new(parent_path).unwrap());
        let cwd_path = Tree::cwd_path().expect("could not canonicalize cwd path");
        self.fs
            .entry(cwd_path.clone())
            .or_insert_with(|| Dir::new(cwd_path.clone()).unwrap());
        let cwd = self.fs.get_mut(&cwd_path).unwrap();
        let idx = cwd.index_by_name(
            init_parent_path
                .file_name()
                .unwrap()
                .to_os_string()
                .into_string()
                .unwrap(),
            state.show_hidden,
        );
        cwd.state.select(Some(idx));
    }

    pub fn cd_selected(&mut self, state: &State) {
        if let Some(selected) = self
            .cwd()
            .get_selected_file(state.show_hidden, &state.query_string)
        {
            if selected.is_dir {
                env::set_current_dir(Path::new(&selected.name)).unwrap();
                let cwd_path =
                    fs::canonicalize(Path::new(".")).expect("could not canonicalize cwd path");

                self.fs
                    .entry(cwd_path.clone())
                    .or_insert_with(|| Dir::new(cwd_path).unwrap());
            }
        }
    }
}
