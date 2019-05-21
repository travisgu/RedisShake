package utils

import (
	"bytes"
	"fmt"
	"strconv"
	"redis-shake/configure"
)

type ClusterNodeInfo struct {
	Id          string
	Address     string
	Flags       string
	Master      string
	PingSent    string
	PongRecv    string
	ConfigEpoch string
	LinkStat    string
	Slot        string
}

func ParseKeyspace(content []byte) (map[int32]int64, error) {
	if bytes.HasPrefix(content, []byte("# Keyspace")) == false {
		return nil, fmt.Errorf("invalid info Keyspace: %s", string(content))
	}

	lines := bytes.Split(content, []byte("\n"))
	reply := make(map[int32]int64)
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if bytes.HasPrefix(line, []byte("db")) == true {
			// line "db0:keys=18,expires=0,avg_ttl=0"
			items := bytes.Split(line, []byte(":"))
			db, err := strconv.Atoi(string(items[0][2:]))
			if err != nil {
				return nil, err
			}
			nums := bytes.Split(items[1], []byte(","))
			if bytes.HasPrefix(nums[0], []byte("keys=")) == false {
				return nil, fmt.Errorf("invalid info Keyspace: %s", string(content))
			}
			keysNum, err := strconv.ParseInt(string(nums[0][5:]), 10, 0)
			if err != nil {
				return nil, err
			}
			reply[int32(db)] = int64(keysNum)
		} // end true
	} // end for
	return reply, nil
}

/*
 * 10.1.1.1:21331> cluster nodes
 * d49a4c7b516b8da222d46a0a589b77f381285977 10.1.1.1:21333@31333 master - 0 1557996786000 3 connected 10923-16383
 * f23ba7be501b2dcd4d6eeabd2d25551513e5c186 10.1.1.1:21336@31336 slave d49a4c7b516b8da222d46a0a589b77f381285977 0 1557996785000 6 connected
 * 75fffcd521738606a919607a7ddd52bcd6d65aa8 10.1.1.1:21331@31331 myself,master - 0 1557996784000 1 connected 0-5460
 * da3dd51bb9cb5803d99942e0f875bc5f36dc3d10 10.1.1.1:21332@31332 master - 0 1557996786260 2 connected 5461-10922
 * eff4e654d3cc361a8ec63640812e394a8deac3d6 10.1.1.1:21335@31335 slave da3dd51bb9cb5803d99942e0f875bc5f36dc3d10 0 1557996787261 5 connected
 * 486e081f8d47968df6a7e43ef9d3ba93b77d03b2 10.1.1.1:21334@31334 slave 75fffcd521738606a919607a7ddd52bcd6d65aa8 0 1557996785258 4 connected
 */
func ParseClusterNode(content []byte) []*ClusterNodeInfo {
	lines := bytes.Split(content, []byte("\r\n"))
	ret := make([]*ClusterNodeInfo, 0, len(lines))
	for _, line := range lines {
		items := bytes.Split(line, []byte(" "))

		address := bytes.Split(items[1], []byte{'@'})
		flag := bytes.Split(items[2], []byte{','})
		var role string
		if len(flag) > 1 {
			role = string(flag[1])
		} else {
			role = string(flag[0])
		}
		var slot string
		if len(items) > 7 {
			slot = string(items[7])
		}
		ret = append(ret, &ClusterNodeInfo{
			Id:          string(items[0]),
			Address:     string(address[0]),
			Flags:       role,
			Master:      string(items[3]),
			PingSent:    string(items[4]),
			PongRecv:    string(items[5]),
			ConfigEpoch: string(items[6]),
			LinkStat:    string(items[7]),
			Slot:        slot,
		})
	}
	return ret
}

// needMaster: true(master), false(slave)
func ClusterNodeChoose(input []*ClusterNodeInfo, needMaster bool) []*ClusterNodeInfo {
	ret := make([]*ClusterNodeInfo, 0, len(input) / 2)
	for _, ele := range input {
		if ele.Flags == conf.StandAloneRoleMaster && needMaster {
			ret = append(ret, ele)
		} else if ele.Flags == conf.StandAloneRoleSlave && !needMaster {
			ret = append(ret, ele)
		}
	}
	return ret
}