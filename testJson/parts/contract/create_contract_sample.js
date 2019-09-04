/*
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { FileSystemWallet, Gateway } = require('fabric-network');
const fs = require('fs');
const path = require('path');
const ccpPath = path.resolve(__dirname, '..', '..','..','fabric-samples', 'basic-network', 'connection.json');
const ccpJSON = fs.readFileSync(ccpPath, 'utf8');
const ccp = JSON.parse(ccpJSON);
var ccControl = {
    create_contract : async function(){
        try {
            // Create a new file system based wallet for managing identities.
            const walletPath = path.join(process.cwd(),'..','..','wallet');
            const wallet = new FileSystemWallet(walletPath);

            // Check to see if we've already enrolled the user.
            const userExists = await wallet.exists('user1');
            if (!userExists) {
                console.log('An identity for the user "user1" does not exist in the wallet');
                console.log('Run the registerUser.js application before retrying');
                return;
            }

            // **************** LOOK HERE **********************
            // Create a new gateway for connecting to our peer node.
            const gateway = new Gateway();
            await gateway.connect(ccp, { wallet, identity: 'user1', discovery: { enabled: false } });

            // Get the network (channel) our contract is deployed to.
            const network = await gateway.getNetwork('mychannel');

            // Get the contract from the network.
            const contract = network.getContract('main');

            let rawdata = fs.readFileSync('./json/IB_CHNCT_KORKF.json');
            let test_json = JSON.parse(rawdata);
            
            const result = await contract.submitTransaction('blockInsert', 'contract', JSON.stringify(test_json));
            // Disconnect from the gateway.
            await gateway.disconnect();
            console.log(result.toString());

        } catch (error) {
            console.error(`Failed to submit transaction: ${error}`);
            process.exit(1);
        }
    }
};


ccControl.create_contract();




