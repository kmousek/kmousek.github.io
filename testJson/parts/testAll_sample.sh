cd ..
echo "  >> admin, user 생성"
node index.js
cd parts
cd contract
echo "  >> 신규 계약 생성"
node create_contract_sample.js
echo "  >> agree contract"
node agree_contract_sample.js
echo "  >> expiration contract"
node expiration_contract_sample.js
echo "  >> upgrade contract sample"
node upgrade_contract_sample.js
echo "  >> agree upgrade contract"
node agree_upgrade_contract_sample.js
echo "  >> contract 조회"
node query_contract_sample.js
cd ..
cd tap
echo "  >> tap input"
node tap_input_sample.js
echo "  >> invoice 발행"
cd ..
cd invoice
node  invoice_insert_sample.js
echo " >> invoice 지급 완료"
node setState_sample.js
echo " >> invoice 조회"
node query_invoice_all.js
