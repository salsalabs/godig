#! /bin/bash
rm -f addressfixer*.log
go run cmd/addressfixer/app/app.go \
--credentials credentials/ewg.yaml \
--criteria 'Email%20IS%20NOT%20EMPTY%26condition%3DEmail%20LIKE%20%25%40%25.%25%26condition%3DReceive_Email%20IN%201%2C2%2C3%2C10%2C15%26condition%3DState%20IS%20EMPTY' \
--fixer-count 25 \
--reader-count 25 \
--file-log
