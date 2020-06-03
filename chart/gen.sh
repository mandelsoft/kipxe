#!/bin/bash

dir="$(realpath "$(dirname "$0")")"

: ${cmd:=merge}
mkdir -p manifests/kipxe
cd manifests


echo "*** $cmd"
case $cmd in
   merge) helm template --name kipxe --namespace metal --values ../values.yaml --output-dir . "$dir"
          mv kipxe/templates/* .;;
          
   diff)  
          for i in ??-*.yaml; do
            echo "$i"
            spiff diff ../metal/"$i"  "$i"
          done;;
esac
echo "final manifests in $(pwd)"
