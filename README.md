# fm

A tui file manager

![fm-preview](https://taybart-samples.nyc3.digitaloceanspaces.com/fm.gif?id=1)

## Commands

|    cmd      | alt cmd | Description                                                                                              |
|-------------|---------|----------------------------------------------------------------------------------------------------------|
| :edit         | ee      | Open active file with $EDITOR                                                                          |
| :inspect      | i       | Open active file with $EDITOR, if editor is vim||nvim. fm will source $CONFIG/vimrc.preview in RO mode |
| :shell        | eS      | Start $SHELL at current directory                                                                      |
| :delete       | ed      | Moves file to temporary location. After fm is closed, the files will be deleted permanently            |
| :undo         | eu      | Put files back where they were and don't delete them at the end of the fm session.                     |
| :toggleHidden | zh      | Flips between hiding and showing hidden files.                                                         |
| :rn newname   |         | Rename active file                                                                                     |
| :!cmd args    | s       | Run command with $SHELL                                                                                |
| :yank         | yy      | Copy file under cursor                                                                                 |
| :cut          | dd      | Cut file under cursor                                                                                  |
| :paste        | pp      | Paste file to current directory                                                                        |
|               | /       | Open fzf and search through current directory select output or cd to output                            |

## Config

Inital config looks like:

```json
{
  "columnWidth": -1,
  "columnRatios": [2,3,5],
  "jumpAmount": 5,
  "previewBlacklist": "*.mp4",
  "folder": "~/.config/fm"
}

```
