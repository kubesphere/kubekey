#!/bin/bash
#count=$(docker-compose -f /opt/harbor/docker-compose.yml ps -a|grep healthy|wc -l)
# 不能频繁调用docker-compose 否则会有非常多的临时目录被创建：/tmp/_MEI*
count=$(docker ps |grep goharbor|grep healthy|wc -l)
status=$(ss -tlnp|grep -w 443|wc -l)
if [ $count -ne 11 -a  ];then
   exit 8
elif [ $status -lt 2 ];then
   exit 9
else
   exit 0
fi
