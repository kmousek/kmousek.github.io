export FILE_NM=$1
export INPUT_FILE_NM=$2
export REC_CNT=$3


echo $FILE_NM
echo $INPUT_FILE_NM


cat tap_hd.json > $FILE_NM
head -$3 $INPUT_FILE_NM  >> $FILE_NM
cat tap_tail.json >> $FILE_NM
