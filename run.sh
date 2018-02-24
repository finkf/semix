#!/bin/sh

dir=$1
shift;

while true; do
		echo "$@"
		$@ &
		pid=$!
		# echo "waiting $dir"
		inotifywait -e modify -r $dir
		kill $pid
done
