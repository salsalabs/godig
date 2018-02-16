#!/bin/bash
cd overlap_details
for i in *-*.csv
do
	left=`echo $i | sed 's/-.*//g'`.csv
	right=`echo $i | sed 's/^[0-9]*-//'| sed 's/.csv//'`.csv
	diff -y -w --suppress-common-lines $left $right | grep -v -P '[\<\|\>]+'
done 
