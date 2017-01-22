echo "hello `date`" | mailx -S smtp=localhost:1025 -s "Hi at `date`" abc@phantomail.com
