echo 1000 mb/s Bandwidth
./sim -stats bw_1.csv -read_speed 1 -write_speed 1
echo 2000 mb/s
./sim -stats bw_2.csv -read_speed 2 -write_speed 2
echo 4000 mb/s
./sim -stats bw_4.csv -read_speed 4 -write_speed 4
echo 8000 mb/s
./sim -stats bw_8.csv -read_speed 8 -write_speed 8
echo 16000 mb/s
./sim -stats bw_16.csv -read_speed 16 -write_speed 16
echo 32000 mb/s
./sim -stats bw_32.csv -read_speed 32 -write_speed 32
echo 64000 mb/s
./sim -stats bw_64.csv -read_speed 64 -write_speed 64
echo 128000 mb/s
./sim -stats bw_128.csv -read_speed 128 -write_speed 128
rm -f warm.csv
mv *.csv stats/bandwidth/
