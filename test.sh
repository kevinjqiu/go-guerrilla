echo "hello `date`" | mailx -S smtp=localhost:25 -s "Hi at `date`" abc@phantomail.com
