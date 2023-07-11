# Visualization

## Development Environment configuration

```bash
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
```

## MP0

we want to track two metrics:

Delay from the time the event is generated to the time it shows up in the centralized logger
The amount of bandwidth used by the centralized logger

For the delay, you can just use the difference between the current time when you are about to print the event and the timestamp of the event itself. For measuring the bandwidth, you will need to track the length of all the messages received by the logger.

You should produce a graph of these two metrics over time.
For the bandwidth,
you should track the average bandwidth across each second of the experiment.
For the delay,
for each second you should plot the
minimum, maximum, median, and 90th percentile delay at each second.
Make sure your graphs and axes are well labeled, with units.


## MP1

Record bandwith in json, named after xxx_b.json

Record delay in json, named after xxx_d.json


```json
{ "timestamp": "", "delay": "", "size": "" }
```


### Commands for accessing and running MP1

Parse raw log files to bandwith and delay logs
```python
python log_parser.py <directory to log files>
```

Generate graphs of bandwidth
```python
python report_bandwidth.py <directory to log files>
```

Generate graphs of delay
```python
python report_delay.py <directory to log files>
```

## Additonal Instructions to Access VMs
```cmd
ssh zonglin7@fa21-cs425-g03-01.cs.illinois.edu
ssh zonglin7@fa21-cs425-g03-02.cs.illinois.edu
ssh zonglin7@fa21-cs425-g03-03.cs.illinois.edu

zonglin7@fa21-cs425-g03-01.cs.illinois.edu:/home/zonglin7/cs425-mps
zonglin7@fa21-cs425-g03-02.cs.illinois.edu:/home/zonglin7/cs425-mps
zonglin7@fa21-cs425-g03-03.cs.illinois.edu:/home/zonglin7/cs425-mps

scp zonglin7@fa21-cs425-g03-02.cs.illinois.edu:/tmp/a.log .
scp zonglin7@fa21-cs425-g03-02.cs.illinois.edu:/tmp/b.log .
scp zonglin7@fa21-cs425-g03-03.cs.illinois.edu:/tmp/c.log .

git clone https://github.com/zonglinpeng/distributed_algorithms-mps.git

cat a.log | grep  "metrics.bandwith" > a_bandwith.log

python3 -u ./script/unix/mp1/gentx.py 0.5 | METRICS=y LOG=json ./mp1-linux-amd64 A 8080 ./lib/mp1/config/vm_testing/3/config_a.txt 2> /tmp/a.log
python3 -u ./script/unix/mp1/gentx.py 0.5 | METRICS=y LOG=json ./mp1-linux-amd64 B 8081 ./lib/mp1/config/vm_testing/3/config_b.txt 2> /tmp/b.log
python3 -u ./script/unix/mp1/gentx.py 0.5 | METRICS=y LOG=json ./mp1-linux-amd64 C 8082 ./lib/mp1/config/vm_testing/3/config_c.txt 2> /tmp/c.log


python3 -u ./script/unix/mp1/gentx.py 5 | METRICS=y LOG=json ./mp1-linux-amd64 node1 8080 ./lib/mp1/config/vm_testing/8/config_1.txt 2> /tmp/1.log
python3 -u ./script/unix/mp1/gentx.py 5 | METRICS=y LOG=json ./mp1-linux-amd64 node2 8081 ./lib/mp1/config/vm_testing/8/config_2.txt 2> /tmp/2.log
python3 -u ./script/unix/mp1/gentx.py 5 | METRICS=y LOG=json ./mp1-linux-amd64 node3 8082 ./lib/mp1/config/vm_testing/8/config_3.txt 2> /tmp/3.log

python3 -u ./script/unix/mp1/gentx.py 5 | METRICS=y LOG=json ./mp1-linux-amd64 node4 8080 ./lib/mp1/config/vm_testing/8/config_4.txt 2> /tmp/4.log
python3 -u ./script/unix/mp1/gentx.py 5 | METRICS=y LOG=json ./mp1-linux-amd64 node5 8081 ./lib/mp1/config/vm_testing/8/config_5.txt 2> /tmp/5.log
python3 -u ./script/unix/mp1/gentx.py 5 | METRICS=y LOG=json ./mp1-linux-amd64 node6 8082 ./lib/mp1/config/vm_testing/8/config_6.txt 2> /tmp/6.log

python3 -u ./script/unix/mp1/gentx.py 5 | METRICS=y LOG=json ./mp1-linux-amd64 node7 8080 ./lib/mp1/config/vm_testing/8/config_7.txt 2> /tmp/7.log
python3 -u ./script/unix/mp1/gentx.py 5 | METRICS=y LOG=json ./mp1-linux-amd64 node8 8081 ./lib/mp1/config/vm_testing/8/config_8.txt 2> /tmp/8.log
```
