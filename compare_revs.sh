#!/usr/bin/env bash

# Copy project to temp directory and then reset it to HEAD to capture output of last commited version
# Then go back to current dir, capture output of current working tree
# Compare outputs for several tasks, notify if differ
# Caveats: this test uses real websites and parsing tasks - so it's not idempotent.
# I should think about better solution

set -e

old_dir=$(mktemp -d)
cur_dir=$(pwd)
task_dir=$cur_dir/test_tasks

trap "echo cleaning up && rm -rf $old_dir && echo done" EXIT

echo "Copying project to $old_dir"
time rsync -ar --exclude "node_modules" $cur_dir/ $old_dir
cd $old_dir
git reset --hard HEAD
cd -

failed=0

for task in $task_dir/*; do
  echo "Task $task"
  old_out=$(mktemp)
  echo "Old version output: $old_out"
  cur_out=$(mktemp)
  echo "Cur version output: $cur_out"

  set +e
    cd $old_dir
    rm -f $old_dir/screenshot.png
    sleep 2
    go run github.com/egor3f/rssalchemy/cmd/extractor -o $old_out "$task"
    if [ $? != 0 ]; then
      echo "Failed to run old version"
      cat $old_out
      exit 1
    fi
    cd -
    sleep 2
    go run github.com/egor3f/rssalchemy/cmd/extractor -o $cur_out "$task"
    if [ $? != 0 ]; then
      echo "Failed to run new version"
      cat $cur_out
      exit 1
    fi
  set -e

  if [ "$(cat $old_out)" != "$(cat $cur_out)" ]; then
    echo "Output differ for $task. To inspect use: "
    echo "diff -u $old_out $cur_out"
    failed=$((failed + 1))
    if [ -f $old_dir/screenshot.png ]; then
      cp $old_dir/screenshot.png $cur_dir/screenshot_old.png
      echo Screenshot of old version output copied to cwd
    fi
  fi
done

echo "-----------"
total=$(ls -1q $task_dir/* | wc -l)
echo "Failed: $failed of $total"

if [ $failed > 0 ]; then
  exit 1
fi
