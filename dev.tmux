#!/usr/bin/env bash

# DEFAULT_PORT
DEFAULT_PORT=80898

# Detect if we are on a Mac
if [[ "$OSTYPE" == "darwin"* ]]; then
  SHELL_CMD="bash"
else
  SHELL_CMD="bash"
fi

# Generate random Docker-like name
generate_session_name() {
  ADJECTIVES=(happy brave calm eager fancy gentle jolly kind lucky nice proud quick shiny witty bold clever)
  NOUNS=(tiger panda eagle otter fox whale lion falcon wolf bear koala shark)

  adj=${ADJECTIVES[$RANDOM % ${#ADJECTIVES[@]}]}
  noun=${NOUNS[$RANDOM % ${#NOUNS[@]}]}
  echo "${adj}-${noun}-dev"
}

SESSION=${2:-$(generate_session_name)}
echo "Using SESSION=$SESSION"

# Check if the session already exists
tmux has-session -t "$SESSION" 2>/dev/null

# Create new sessions [Optional: ZSH default shell]
if [ $? != 0 ]; then
  tmux new-session -d -s "$SESSION" "$SHELL_CMD"
  tmux split-window -h -t "$SESSION" "$SHELL_CMD"
  tmux split-window -v -t "$SESSION" "$SHELL_CMD"

  # Optional: Send terminal command to a pane if needed
  # tmux send-keys -t "$SESSION":0.0 "ls" C-m
  #
  # Send the hugo command to pane 0
  # Example: C-m at the end simulates pressing "Enter"
  # tmux send-keys -t $SESSION:0 "until hugo serve -D -p $PORT; do echo 'Command failed. Press Enter to retry...'; read; done" C-m

  tmux send-keys -t "$SESSION":0


fi

tmux attach-session -t "$SESSION"
