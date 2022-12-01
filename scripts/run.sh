#!/bin/bash

export NEW_TAG=`echo $IMAGE_TAG_FULL | cut -d ":" -f 2`
python krm_release.py $FUNCTION_NAME $NEW_TAG 
