# fs-monitor
golang 기반 inotify 을 활용한 파일시스템 모니터링


리눅스의 경우 /proc/sys/fs/inotify/max_user_watches 값에 따라 파일개수가 제한 되는 듯하고
최대치는 sysctl 로 계정당 10만개까지 가능 한걸로 보임.

대용량 파일 시스템을 모니터링 할때는 직접 해당 기능을 구현하는 편이 나을듯. 메모리에 올려서 주기적으로 변화를 체크 하고
알럿하는 형식으로..
