#! /bin/bash
rm addressfixer*.log
go run cmd/addressfixer/app/app.go \
--credentials credentials/ewg.yaml \
--criteria 'Email%20IS%20NOT%20EMPTY&Receive_Email>0&State%20IS%20EMPTY' \
--fixer-count 25 \
--reader-count 25 \
--file-log
