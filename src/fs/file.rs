use std::fs::{metadata, DirEntry};

// use tui::{
//     style::{Color, Modifier, Style},
//     widgets::{List, ListItem, ListState},
// };

pub struct File {
    pub name: String,
    pub is_dir: bool,
    pub is_hidden: bool,
    // pub metadata: Metadata,
}

impl File {
    pub fn new(file: DirEntry) -> File {
        let path = file.path();
        let metadata = metadata(path.clone()).unwrap();

        // we just want a displayable string
        let name = path
            .file_name()
            .unwrap()
            .to_owned()
            .to_str()
            .unwrap()
            .to_owned();

        let is_hidden = name.clone().starts_with('.');
        File {
            name,
            is_dir: metadata.is_dir(),
            is_hidden,
        }
    }
    pub fn get_contents(&self) -> String {
        if self.is_dir {
            return "".to_string();
        }
        "contents".to_string()
    }
}
