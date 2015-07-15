#!/bin/bash
mkdir originals
shopt -s nocaseglob # case insensitive pattern matching
cp [0-9]*jpg originals/
touch a.JPG
for i in [0-9]*jpg; do
  convert -crop 1600x800+1100+1600 +repage "$i" a.JPG
  mv a.JPG "$i" 
done
rm a.JPG
shopt -u nocaseglob # case sensitive pattern matching
