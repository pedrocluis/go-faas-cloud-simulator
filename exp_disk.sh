echo 50 Disk
./sim -stats disk_50.csv -disk 50000
echo 100 Disk
./sim -stats disk_100.csv -disk 100000
echo 250 Disk
./sim -stats disk_250.csv -disk 250000
echo 500 Disk
./sim -stats disk_500.csv -disk 500000
echo 1000 Disk
./sim -stats disk_1000.csv -disk 1000000
rm -f warm.csv
mv *.csv stats/disk/
