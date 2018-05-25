#! /bin/bash
rm addressfixer*-
go run cmd/addressfixer/app/app.go \
--credentials credentials/ewg.yaml \
--criteria 'Email%20IS%20NOT%20EMPTY&Receive_Email>0&State%20IS%20EMPTY' \
--fixer-count 10 \
--reader-count 10 \
--file-log
