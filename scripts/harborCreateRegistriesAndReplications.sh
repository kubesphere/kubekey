#!/bin/bash


function createRegistries() {
  
  # create registry     	
  curl -k -u $Harbor_UserPwd  -X POST -H "Content-Type: application/json" "https://${Harbor_master1_Address}/api/v2.0/registries" -d "{\"name\": \"master1_2_master2\", \"type\": \"harbor\", \"url\":\"https://${master2_Address}:7443\", \"credential\": {\"access_key\": \"${Harbor_User}\", \"access_secret\": \"${Harbor_Passwd}\"}, \"insecure\": true}"
  # create registry     	
  curl -k -u $Harbor_UserPwd  -X POST -H "Content-Type: application/json" "https://${Harbor_master1_Address}/api/v2.0/registries" -d "{\"name\": \"master1_2_master3\", \"type\": \"harbor\", \"url\":\"https://${master3_Address}:7443\", \"credential\": {\"access_key\": \"${Harbor_User}\", \"access_secret\": \"${Harbor_Passwd}\"}, \"insecure\": true}"

  # create registry     	
  curl -k -u $Harbor_UserPwd  -X POST -H "Content-Type: application/json" "https://${Harbor_master2_Address}/api/v2.0/registries" -d "{\"name\": \"master2_2_master1\", \"type\": \"harbor\", \"url\":\"https://${master1_Address}:7443\", \"credential\": {\"access_key\": \"${Harbor_User}\", \"access_secret\": \"${Harbor_Passwd}\"}, \"insecure\": true}"
  # create registry     	
  curl -k -u $Harbor_UserPwd  -X POST -H "Content-Type: application/json" "https://${Harbor_master2_Address}/api/v2.0/registries" -d "{\"name\": \"master2_2_master3\", \"type\": \"harbor\", \"url\":\"https://${master3_Address}:7443\", \"credential\": {\"access_key\": \"${Harbor_User}\", \"access_secret\": \"${Harbor_Passwd}\"}, \"insecure\": true}"

  # create registry     	
  curl -k -u $Harbor_UserPwd  -X POST -H "Content-Type: application/json" "https://${Harbor_master3_Address}/api/v2.0/registries" -d "{\"name\": \"master3_2_master1\", \"type\": \"harbor\", \"url\":\"https://${master1_Address}:7443\", \"credential\": {\"access_key\": \"${Harbor_User}\", \"access_secret\": \"${Harbor_Passwd}\"}, \"insecure\": true}"
  # create registry     	
  curl -k -u $Harbor_UserPwd  -X POST -H "Content-Type: application/json" "https://${Harbor_master3_Address}/api/v2.0/registries" -d "{\"name\": \"master3_2_master2\", \"type\": \"harbor\", \"url\":\"https://${master2_Address}:7443\", \"credential\": {\"access_key\": \"${Harbor_User}\", \"access_secret\": \"${Harbor_Passwd}\"}, \"insecure\": true}"

}

function listRegistries() {
  curl -k -u $Harbor_UserPwd  -X GET -H "Content-Type: application/json" "https://${Harbor_master1_Address}/api/v2.0/registries"
  curl -k -u $Harbor_UserPwd  -X GET -H "Content-Type: application/json" "https://${Harbor_master2_Address}/api/v2.0/registries"
  curl -k -u $Harbor_UserPwd  -X GET -H "Content-Type: application/json" "https://${Harbor_master3_Address}/api/v2.0/registries"

}

function createReplication() {

  curl -k -u $Harbor_UserPwd  -X POST -H "Content-Type: application/json" "https://${Harbor_master1_Address}/api/v2.0/replication/policies" -d "{\"name\": \"master1_2_master2\", \"enabled\": true, \"deletion\":true, \"override\":true, \"replicate_deletion\":true, \"dest_registry\":{ \"id\": 1, \"name\": \"master1_2_master2\"}, \"trigger\": {\"type\": \"event_based\"}, \"dest_namespace_replace_count\":1 }"
  curl -k -u $Harbor_UserPwd  -X POST -H "Content-Type: application/json" "https://${Harbor_master1_Address}/api/v2.0/replication/policies" -d "{\"name\": \"master1_2_master3\", \"enabled\": true, \"deletion\":true, \"override\":true, \"replicate_deletion\":true, \"dest_registry\":{ \"id\": 2, \"name\": \"master1_2_master3\"}, \"trigger\": {\"type\": \"event_based\"}, \"dest_namespace_replace_count\":1 }"

  curl -k -u $Harbor_UserPwd  -X POST -H "Content-Type: application/json" "https://${Harbor_master2_Address}/api/v2.0/replication/policies" -d "{\"name\": \"master2_2_master1\", \"enabled\": true, \"deletion\":true, \"override\":true, \"replicate_deletion\":true, \"dest_registry\":{ \"id\": 1, \"name\": \"master2_2_master1\"}, \"trigger\": {\"type\": \"event_based\"}, \"dest_namespace_replace_count\":1 }"
  curl -k -u $Harbor_UserPwd  -X POST -H "Content-Type: application/json" "https://${Harbor_master2_Address}/api/v2.0/replication/policies" -d "{\"name\": \"master2_2_master3\", \"enabled\": true, \"deletion\":true, \"override\":true, \"replicate_deletion\":true, \"dest_registry\":{ \"id\": 2, \"name\": \"master2_2_master3\"}, \"trigger\": {\"type\": \"event_based\"}, \"dest_namespace_replace_count\":1 }"

  curl -k -u $Harbor_UserPwd  -X POST -H "Content-Type: application/json" "https://${Harbor_master3_Address}/api/v2.0/replication/policies" -d "{\"name\": \"master3_2_master1\", \"enabled\": true, \"deletion\":true, \"override\":true, \"replicate_deletion\":true, \"dest_registry\":{ \"id\": 1, \"name\": \"master3_2_master1\"}, \"trigger\": {\"type\": \"event_based\"}, \"dest_namespace_replace_count\":1 }"
  curl -k -u $Harbor_UserPwd  -X POST -H "Content-Type: application/json" "https://${Harbor_master3_Address}/api/v2.0/replication/policies" -d "{\"name\": \"master3_2_master2\", \"enabled\": true, \"deletion\":true, \"override\":true, \"replicate_deletion\":true, \"dest_registry\":{ \"id\": 2, \"name\": \"master3_2_master2\"}, \"trigger\": {\"type\": \"event_based\"}, \"dest_namespace_replace_count\":1 }"
}

function listReplications() {

  curl -k -u $Harbor_UserPwd  -X GET -H "Content-Type: application/json" "https://${Harbor_master1_Address}/api/v2.0/replication/policies"
  curl -k -u $Harbor_UserPwd  -X GET -H "Content-Type: application/json" "https://${Harbor_master2_Address}/api/v2.0/replication/policies"
  curl -k -u $Harbor_UserPwd  -X GET -H "Content-Type: application/json" "https://${Harbor_master3_Address}/api/v2.0/replication/policies"
}

####  main ######
Harbor_master1_Address=master1:7443
master1_Address=192.168.122.61
Harbor_master2_Address=master2:7443
master2_Address=192.168.122.62
Harbor_master3_Address=master3:7443
master3_Address=192.168.122.63
Harbor_User=admin                                  #登录Harbor的用户
Harbor_Passwd="Harbor12345"                  #登录Harbor的用户密码
Harbor_UserPwd="$Harbor_User:$Harbor_Passwd"


createRegistries
listRegistries
createReplication
listReplications
