# Display tmux shortcuts in the console

Depending on screen width, the content will be reordered to take as much space as possible.

To automatically execute this in a new tmux session, add the generated binary to the path, and then append the following in `~/.tmux.conf`:

```
set-option -g default-command "shortcuts; $SHELL"
```
