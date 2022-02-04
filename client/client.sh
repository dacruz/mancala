#!/usr/bin/env sh

cookies=`echo cookies-$RANDOM`  
URL=$1
match_id=`curl -c $cookies $URL 2>>/dev/null | jq '.match'| cut -d '"' -f2`

while true; do
    match=`curl -b $cookies $URL/$match_id 2>>/dev/null`

    my_turn=`echo $match | jq '.my_turn'`
    clear
    
    my_board="`echo $match | jq -c '.pits[0][0:6]'`"
    my_big_pit="`echo $match | jq -c '.pits[0][6]'`"
    opponents_board=`echo $match | jq -c '.pits[1][0:6]'` 
    opponents_big_pit="`echo $match | jq -c '.pits[1][6]'`"

    echo "           |       pits        | big pit "
    echo "_________________________________________"
    echo "opponent   |   $opponents_board   |   $opponents_big_pit   |"
    echo "you        |   $my_board   |   $my_big_pit   |"
    

    if "$my_turn" == "true"; then          
        echo "Select pit number from 0 to 5"
        read pit  
        curl -X POST -b $cookies $URL/$match_id/$pit
    else 
        echo "Waiting for opponent"
    fi
    sleep 1
done
