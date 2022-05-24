#!/bin/bash

rm snapshots/private_tangle1/full_snapshot.bin
rm snapshots/private_tangle1/delta_snapshot.bin
rm snapshots/private_tangle2/full_snapshot.bin
rm snapshots/private_tangle2/delta_snapshot.bin
rm snapshots/private_tangle3/full_snapshot.bin
rm snapshots/private_tangle3/delta_snapshot.bin
rm snapshots/private_tangle4/full_snapshot.bin
rm snapshots/private_tangle4/delta_snapshot.bin
rm snapshots/private_tangle5/full_snapshot.bin
rm snapshots/private_tangle5/delta_snapshot.bin
rm snapshots/private_tangle6/full_snapshot.bin
rm snapshots/private_tangle6/delta_snapshot.bin
rm snapshots/private_tangle7/full_snapshot.bin
rm snapshots/private_tangle7/delta_snapshot.bin
rm snapshots/private_tangle8/full_snapshot.bin
rm snapshots/private_tangle8/delta_snapshot.bin
rm snapshots/private_tangle9/full_snapshot.bin
rm snapshots/private_tangle9/delta_snapshot.bin
rm snapshots/private_tangle10/full_snapshot.bin
rm snapshots/private_tangle10/delta_snapshot.bin
rm snapshots/private_tangle11/full_snapshot.bin
rm snapshots/private_tangle11/delta_snapshot.bin
rm snapshots/private_tangle12/full_snapshot.bin
rm snapshots/private_tangle12/delta_snapshot.bin
rm snapshots/private_tangle13/full_snapshot.bin
rm snapshots/private_tangle13/delta_snapshot.bin
rm snapshots/private_tangle14/full_snapshot.bin
rm snapshots/private_tangle14/delta_snapshot.bin
rm snapshots/private_tangle15/full_snapshot.bin
rm snapshots/private_tangle15/delta_snapshot.bin
rm snapshots/private_tangle16/full_snapshot.bin
rm snapshots/private_tangle16/delta_snapshot.bin
rm snapshots/private_tangle17/full_snapshot.bin
rm snapshots/private_tangle17/delta_snapshot.bin
rm snapshots/private_tangle18/full_snapshot.bin
rm snapshots/private_tangle18/delta_snapshot.bin
rm snapshots/private_tangle19/full_snapshot.bin
rm snapshots/private_tangle19/delta_snapshot.bin
rm snapshots/private_tangle20/full_snapshot.bin
rm snapshots/private_tangle20/delta_snapshot.bin
rm snapshots/private_tangle21/full_snapshot.bin
rm snapshots/private_tangle21/delta_snapshot.bin
rm snapshots/private_tangle22/full_snapshot.bin
rm snapshots/private_tangle22/delta_snapshot.bin
mkdir -p snapshots/private_tangle1/
mkdir -p snapshots/private_tangle2/
mkdir -p snapshots/private_tangle3/
mkdir -p snapshots/private_tangle4/
mkdir -p snapshots/private_tangle5/
mkdir -p snapshots/private_tangle6/
mkdir -p snapshots/private_tangle7/
mkdir -p snapshots/private_tangle8/
mkdir -p snapshots/private_tangle9/
mkdir -p snapshots/private_tangle10/
mkdir -p snapshots/private_tangle11/
mkdir -p snapshots/private_tangle12/
mkdir -p snapshots/private_tangle13/
mkdir -p snapshots/private_tangle14/
mkdir -p snapshots/private_tangle15/
mkdir -p snapshots/private_tangle16/
mkdir -p snapshots/private_tangle17/
mkdir -p snapshots/private_tangle18/
mkdir -p snapshots/private_tangle19/
mkdir -p snapshots/private_tangle20/
mkdir -p snapshots/private_tangle21/
mkdir -p snapshots/private_tangle22/
go run ../../main.go tool snap-gen private_tangle1 60200bad8137a704216e84f8f9acfe65b972d9f4155becb4815282b03cef99fe 1000000000 snapshots/private_tangle1/full_snapshot.bin
cp snapshots/private_tangle1/full_snapshot.bin snapshots/private_tangle2/
cp snapshots/private_tangle1/full_snapshot.bin snapshots/private_tangle3/
cp snapshots/private_tangle1/full_snapshot.bin snapshots/private_tangle4/
cp snapshots/private_tangle1/full_snapshot.bin snapshots/private_tangle5/
cp snapshots/private_tangle1/full_snapshot.bin snapshots/private_tangle6/
cp snapshots/private_tangle1/full_snapshot.bin snapshots/private_tangle7/
cp snapshots/private_tangle1/full_snapshot.bin snapshots/private_tangle8/
cp snapshots/private_tangle1/full_snapshot.bin snapshots/private_tangle9/
cp snapshots/private_tangle1/full_snapshot.bin snapshots/private_tangle10/
cp snapshots/private_tangle1/full_snapshot.bin snapshots/private_tangle11/
cp snapshots/private_tangle1/full_snapshot.bin snapshots/private_tangle12/
cp snapshots/private_tangle1/full_snapshot.bin snapshots/private_tangle13/
cp snapshots/private_tangle1/full_snapshot.bin snapshots/private_tangle14/
cp snapshots/private_tangle1/full_snapshot.bin snapshots/private_tangle15/
cp snapshots/private_tangle1/full_snapshot.bin snapshots/private_tangle16/
cp snapshots/private_tangle1/full_snapshot.bin snapshots/private_tangle17/
cp snapshots/private_tangle1/full_snapshot.bin snapshots/private_tangle18/
cp snapshots/private_tangle1/full_snapshot.bin snapshots/private_tangle19/
cp snapshots/private_tangle1/full_snapshot.bin snapshots/private_tangle20/
cp snapshots/private_tangle1/full_snapshot.bin snapshots/private_tangle21/
cp snapshots/private_tangle1/full_snapshot.bin snapshots/private_tangle22/


docker-compose up

echo "Clean up docker resources"
docker-compose down -v