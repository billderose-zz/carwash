#!/bin/bash
mkdir originals
cp *.JPG originals
touch a.JPG
shopt -s nocaseglob # case insensitive pattern matching
for i in *.JPG; do
  convert -crop 1600x800+1100+1600 +repage "$i" a.JPG
  mv a.JPG "$i" 
done
rm a.JPG
shopt -u nocaseglob # case sensitive pattern matching
