#!/bin/bash
shopt -s nocaseglob # case insensitive pattern matching
a=$(ls [0-9]*jpg | sort -rn | awk '{printf "%d", $1 + 1; exit}') # find the highest numbered file
for i in *.jpg; do
  cmp=${i%%.*} # drop the file extension
  if [ "$cmp" -lt "$a" ]
  then
      continue
  else
      new=$(printf "%d.JPG" "$a")
      mv "$i" "$new"
      let a=a+1
  fi
done
shopt -u nocaseglob # case sensitive pattern matching