#!/bin/bash

# total=$(free -m | awk '/Mem:/ {print $2}'); used=$(free -m | awk '/Mem:/ {print $3}'); free=$(free -m | awk '/Mem:/ {print $4}'); available=$(free -m | awk '/Mem:/ {print $7}'); cached=$(vmstat -s | grep "cached memory" | awk '{print int($1/1024)}'); swap_total=$(free -m | awk '/Swap:/ {print $2}'); swap_used=$(free -m | awk '/Swap:/ {print $3}'); video_wait=$(vmstat 1 2 | tail -1 | awk '{print $16}'); used_pct=$((total>0 ? 100*used/total : 0)); free_pct=$((total>0 ? 100*free/total : 0)); cached_pct=$((total>0 ? 100*cached/total : 0)); avail_pct=$((total>0 ? 100*available/total : 0)); swap_pct=$((swap_total>0 ? 100*swap_used/swap_total : 0)); echo "{\"memory_info\": {\"total\": {\"total\": $total, \"total_unit\": \"MB\", \"actual\": $total, \"actual_unit\": \"MB\", \"percent\": 100}, \"used\": {\"total\": $total, \"total_unit\": \"MB\", \"actual\": $used, \"actual_unit\": \"MB\", \"percent\": $used_pct}, \"free\": {\"total\": $total, \"total_unit\": \"MB\", \"actual\": $free, \"actual_unit\": \"MB\", \"percent\": $free_pct}, \"cached\": {\"total\": $total, \"total_unit\": \"MB\", \"actual\": $cached, \"actual_unit\": \"MB\", \"percent\": $cached_pct}, \"available\": {\"total\": $total, \"total_unit\": \"MB\", \"actual\": $available, \"actual_unit\": \"MB\", \"percent\": $avail_pct}, \"swap\": {\"total\": $swap_total, \"total_unit\": \"MB\", \"actual\": $swap_used, \"actual_unit\": \"MB\", \"percent\": $swap_pct}, \"video\": {\"total\": 100, \"total_unit\": \"%\", \"actual\": $video_wait, \"actual_unit\": \"%\", \"percent\": $video_wait}}}"


        total=$(free -m | awk '/Mem:/ {print $2}'); \
        used=$(free -m | awk '/Mem:/ {print $3}'); \
        free=$(free -m | awk '/Mem:/ {print $4}'); \
        available=$(free -m | awk '/Mem:/ {print $7}'); \
        cached=$(vmstat -s | grep "cached memory" | awk '{print int($1/1024)}'); \
        cached=${cached:-0}; \
        swap_total=$(free -m | awk '/Swap:/ {print $2}'); \
        swap_used=$(free -m | awk '/Swap:/ {print $3}'); \
        video_wait=$(vmstat 1 2 | tail -1 | awk '{print $16}'); \
        used_pct=$((total>0 ? 100*used/total : 0)); \
        free_pct=$((total>0 ? 100*free/total : 0)); \
        cached_pct=$((total>0 ? 100*cached/total : 0)); \
        avail_pct=$((total>0 ? 100*available/total : 0)); \
        swap_pct=$((swap_total>0 ? 100*swap_used/swap_total : 0)); \
        \
        echo "{\"memory_info\": {\"total\": {\"total\": $total, \"total_unit\": \"MB\", \"actual\": $total, \"actual_unit\": \"MB\", \"percent\": 100}, \"used\": {\"total\": $total, \"total_unit\": \"MB\", \"actual\": $used, \"actual_unit\": \"MB\", \"percent\": $used_pct}, \"free\": {\"total\": $total, \"total_unit\": \"MB\", \"actual\": $free, \"actual_unit\": \"MB\", \"percent\": $free_pct}, \"cached\": {\"total\": $total, \"total_unit\": \"MB\", \"actual\": ${cached:-0}, \"actual_unit\": \"MB\", \"percent\": $cached_pct}, \"available\": {\"total\": $total, \"total_unit\": \"MB\", \"actual\": $available, \"actual_unit\": \"MB\", \"percent\": $avail_pct}, \"swap\": {\"total\": $swap_total, \"total_unit\": \"MB\", \"actual\": $swap_used, \"actual_unit\": \"MB\", \"percent\": $swap_pct}, \"video\": {\"total\": 100, \"total_unit\": \"%\", \"actual\": $video_wait, \"actual_unit\": \"%\", \"percent\": $video_wait}}}"
