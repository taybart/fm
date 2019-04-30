# fm

A tui file manager written in go. Ranger is written in icky python and it is far too complicated.

![30-4-2019](https://thumbs.gfycat.com/HalfFondImperialeagle-size_restricted.gif)

## Commands

| : cmd       | alt cmd | Function            | Description                                                                                            |
|-------------|---------|---------------------|--------------------------------------------------------------------------------------------------------|
| e           | ee      | Edit                | Open active file with $EDITOR                                                                          |
| i           | i       | Inspect             | Open active file with $EDITOR, if editor is vim||nvim. fm will source $CONFIG/vimrc.preview in RO mode |
| sh          | eS      | Shell               | Start $SHELL at current directory                                                                      |
| d           | ed      | Soft Delete         | Moves file to temporary location. After fm is closed, the files will be deleted permanently            |
| ud          | eu      | Undo Delete.        | Put files back where they were and don't delete them at the send of the fm session.                    |
| D           | eD      | Permanent Delete    | Remove files right away and don't move them. (Suggested for big files)                                 |
| h           | eh      | Toggle hidden files | Flips between hiding and showing hidden files.                                                         |
| rn newname  |         | Rename              | Rename active file                                                                                     |
| !cmd args   | s       |                     | Run command with $SHELL                                                                                |
|             | /       | Search              | Open fzf and search through current directory                                                          |

## Config

Inital config looks like:

```json
{
  "showHidden": false,
  "wrapText": true,
  "columnWidth": -1,
  "columnRatios": [2,3,5],
  "jumpAmount": 5,
  "previewRegex": "*",
  "folder": "~/.config/fm"
}

```
