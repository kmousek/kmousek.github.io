cd ..
./teardown.sh

./startFabric.sh

cd ccControl

node index.js

node enrollAdmin.js

node registerUser.js

cd parts
cd contract

node create_contract_sample.js

node agree_contract_sample.js

#node get_active_sample.js

cd ..

cd tap

node tap_input_moclocal.js
node tap_input_mochome.js
node tap_input_mocint.js
node tap_input_mtc.js
node tap_input_smsmo.js
node tap_input_smsmt.js
node tap_input_gprs.js


##sleep 10

##node tap_input_sample_sms.js

##node tap_input_sample_gprs.js
