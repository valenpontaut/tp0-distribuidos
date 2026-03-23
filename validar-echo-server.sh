#!/bin/bash                                                                                                                       

MESSAGE="message from client"                                                                                                            

RESPONSE=$(echo "$MESSAGE" | docker run --rm -i --network tp0_testing_net busybox nc server 12345)                         

if [ "$RESPONSE" = "$MESSAGE" ]; then                                                                                      
  echo "action: test_echo_server | result: success"                                                                      
else                                                                                                                       
  echo "action: test_echo_server | result: fail"                                                                         
fi 