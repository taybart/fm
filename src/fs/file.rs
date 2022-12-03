use crate::finder::match_and_score_with_positions;
use std::fs::{metadata, read_to_string, DirEntry, Metadata};
use std::path::PathBuf;

use tui::{
    style::{Color, Style},
    text::{Span, Spans},
};

#[derive(Clone)]
pub struct File {
    pub name: String,
    pub path: PathBuf,
    pub is_dir: bool,
    pub is_hidden: bool,
    pub metadata: Metadata,
}

impl File {
    pub fn new(file: DirEntry) -> File {
        let path = file.path();
        let metadata = metadata(path.clone()).unwrap();

        let name = file
            .file_name()
            .into_string()
            .expect("convert file os string to str");

        let is_hidden = name.starts_with('.');
        File {
            name,
            path,
            is_dir: metadata.is_dir(),
            is_hidden,
            metadata,
        }
    }

    /// formats the file name with matched letters highlighted
    pub fn display_with_query(&self, query: &str) -> Option<(f64, Spans)> {
        match match_and_score_with_positions(query, &self.name) {
            Some(mut matches) => {
                let mut texts = Vec::new();
                let mut string = "".to_string();
                for (i, c) in matches.1.chars().enumerate() {
                    // is this character a matched character?
                    if matches.2.len() > 0 && i == matches.2[0] {
                        texts.push(Span::raw(string.clone()));
                        string = "".to_string();
                        matches.2.remove(0);
                        texts.push(Span::styled(c.to_string(), Style::default().fg(Color::Red)));

                    // otherwise keep building string
                    } else {
                        string.push(c);
                    }
                }
                // the rest of the string
                texts.push(Span::raw(format!("{} {}", string.clone(), matches.0)));

                // return score and formatted name
                Some((matches.0, Spans::from(texts)))
            }
            None => None,
        }
    }

    pub fn get_contents(&self) -> String {
        read_to_string(&self.name)
            .or::<String>(Ok("".to_string()))
            .unwrap()
    }
}
