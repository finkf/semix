#!/bin/sh

pid=0
cleanup () {
		kill $pid
}
trap cleanup EXIT

if test $# -lt 2 -o ! -e "$1" ; then
		echo "error: usage $0 [dir|file] command...";
		exit 1;
fi

dir=$1
shift;

while true; do
		echo $(date): "starting: $@"
		$@ &
		pid=$!
		inotifywait -e modify -r $dir
		kill $pid
		wait $pid
done
