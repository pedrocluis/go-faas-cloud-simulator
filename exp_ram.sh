echo 2 RAM
./sim -stats ram_2.csv -ram 2000
echo 4 RAM
./sim -stats ram_4.csv -ram 4000
echo 8 RAM
./sim -stats ram_8.csv -ram 8000
echo 16 RAM
./sim -stats ram_16.csv -ram 16000
echo 32 RAM
./sim -stats ram_32.csv -ram 32000
echo 64 RAM
./sim -stats ram_64.csv -ram 64000
echo 128 RAM
./sim -stats ram_128.csv -ram 128000
rm -f lukewarm.csv
mv *.csv stats/ram/
