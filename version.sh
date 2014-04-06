#!/bin/bash

ph=__VERSION__
ver=`printf "%s(%s)" $(<VERSION) $(git rev-parse HEAD)`

if [ $# -ne 1 ];then
  exit -1
elif [ $1 == 'unset' ];then
  regex="s/$ver/$ph/"
elif [ $1 == 'set' ];then
  regex="s/$ph/$ver/"
else
  exit -1
fi

sed -i '' -e "$regex" version.go
