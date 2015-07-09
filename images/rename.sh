#!/bin/bash
shopt -s nocaseglob # case insensitive pattern matching
a=0
for i in *.JPG; do
  new=$(printf "%d.JPG" "$a")
  mv -- "$i" "$new"
  let a=a+1
done
shopt -u nocaseglob # case sensitive pattern matching