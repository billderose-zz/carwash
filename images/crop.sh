#!/bin/bash
shopt -s nocaseglob # case insensitive pattern matching
for i in *.JPG; do
  convert -crop 1600x800+1100+1600 +repage "$i" a.JPG
  mv a.JPG "$i" 
  rm a.JPG
done
shopt -u nocaseglob # case sensitive pattern matching
