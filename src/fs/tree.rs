use crate::fs::dir::Dir;

pub struct Tree {
    pub parent: Dir,
    pub cwd: Dir,
}

impl Tree {
    pub fn new() -> Tree {
        Tree {
            parent: Dir::new(".."),
            cwd: Dir::new("."),
        }
    }
}
