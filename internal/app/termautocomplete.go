package app

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Licoy/sshyard/internal/theme"
	"github.com/Licoy/sshyard/internal/vt"
	"github.com/egoist/mygo/ui"
)

type termSuggestKind uint8

const (
	suggestSnippet termSuggestKind = iota
	suggestLinux
)

type termSuggestion struct {
	Kind termSuggestKind
	Cmd  string
	Desc string
}

const (
	acMaxSuggestions = 12
	acMaxVisible     = 6
)

// linuxCommandCatalog holds common Linux base commands and their subcommands.
var linuxCommandCatalog = []termSuggestion{
	// apt / apt-get
	{Kind: suggestLinux, Cmd: "apt update", Desc: "更新软件包索引"},
	{Kind: suggestLinux, Cmd: "apt upgrade", Desc: "升级已安装软件包"},
	{Kind: suggestLinux, Cmd: "apt full-upgrade", Desc: "完整升级系统软件包"},
	{Kind: suggestLinux, Cmd: "apt install ", Desc: "安装软件包"},
	{Kind: suggestLinux, Cmd: "apt remove ", Desc: "卸载软件包"},
	{Kind: suggestLinux, Cmd: "apt purge ", Desc: "彻底卸载软件包及配置"},
	{Kind: suggestLinux, Cmd: "apt autoremove", Desc: "自动清理无用依赖"},
	{Kind: suggestLinux, Cmd: "apt search ", Desc: "搜索软件包"},
	{Kind: suggestLinux, Cmd: "apt show ", Desc: "查看软件包详情"},
	{Kind: suggestLinux, Cmd: "apt list --installed", Desc: "列出已安装软件包"},
	{Kind: suggestLinux, Cmd: "apt list --upgradable", Desc: "列出可升级软件包"},
	{Kind: suggestLinux, Cmd: "apt clean", Desc: "清理安装包缓存"},
	{Kind: suggestLinux, Cmd: "apt-get update", Desc: "更新软件包索引"},
	{Kind: suggestLinux, Cmd: "apt-get upgrade -y", Desc: "升级已安装软件包"},
	{Kind: suggestLinux, Cmd: "apt-get install -y ", Desc: "安装软件包"},
	{Kind: suggestLinux, Cmd: "apt-get remove ", Desc: "卸载软件包"},
	{Kind: suggestLinux, Cmd: "apt-get autoremove -y", Desc: "自动清理无用依赖"},
	{Kind: suggestLinux, Cmd: "apt-get clean", Desc: "清理下载缓存"},

	// yum / dnf
	{Kind: suggestLinux, Cmd: "yum makecache", Desc: "生成软件包缓存"},
	{Kind: suggestLinux, Cmd: "yum update -y", Desc: "升级所有软件包"},
	{Kind: suggestLinux, Cmd: "yum install -y ", Desc: "安装软件包"},
	{Kind: suggestLinux, Cmd: "yum remove ", Desc: "卸载软件包"},
	{Kind: suggestLinux, Cmd: "yum search ", Desc: "搜索软件包"},
	{Kind: suggestLinux, Cmd: "yum list installed", Desc: "列出已安装软件包"},
	{Kind: suggestLinux, Cmd: "dnf update -y", Desc: "升级所有软件包"},
	{Kind: suggestLinux, Cmd: "dnf install -y ", Desc: "安装软件包"},
	{Kind: suggestLinux, Cmd: "dnf remove ", Desc: "卸载软件包"},

	// systemctl / service / journalctl
	{Kind: suggestLinux, Cmd: "systemctl status ", Desc: "查看服务运行状态"},
	{Kind: suggestLinux, Cmd: "systemctl start ", Desc: "启动服务"},
	{Kind: suggestLinux, Cmd: "systemctl stop ", Desc: "停止服务"},
	{Kind: suggestLinux, Cmd: "systemctl restart ", Desc: "重启服务"},
	{Kind: suggestLinux, Cmd: "systemctl reload ", Desc: "重载服务配置"},
	{Kind: suggestLinux, Cmd: "systemctl enable ", Desc: "设置服务开机自启"},
	{Kind: suggestLinux, Cmd: "systemctl enable --now ", Desc: "设置自启并立即启动"},
	{Kind: suggestLinux, Cmd: "systemctl disable ", Desc: "取消服务开机自启"},
	{Kind: suggestLinux, Cmd: "systemctl daemon-reload", Desc: "重载 systemd 配置"},
	{Kind: suggestLinux, Cmd: "systemctl list-units --type=service --state=running", Desc: "查看运行中的服务"},
	{Kind: suggestLinux, Cmd: "systemctl list-unit-files --type=service", Desc: "查看所有服务自启状态"},
	{Kind: suggestLinux, Cmd: "journalctl -xe", Desc: "查看最近系统错误日志"},
	{Kind: suggestLinux, Cmd: "journalctl -f", Desc: "实时跟踪系统日志"},
	{Kind: suggestLinux, Cmd: "journalctl -u ", Desc: "查看指定服务日志"},
	{Kind: suggestLinux, Cmd: "journalctl --vacuum-time=7d", Desc: "清理 7 天前的日志"},
	{Kind: suggestLinux, Cmd: "journalctl --disk-usage", Desc: "查看日志占用磁盘空间"},

	// docker / docker compose
	{Kind: suggestLinux, Cmd: "docker ps", Desc: "查看运行中的容器"},
	{Kind: suggestLinux, Cmd: "docker ps -a", Desc: "查看所有容器"},
	{Kind: suggestLinux, Cmd: "docker images", Desc: "查看本地镜像列表"},
	{Kind: suggestLinux, Cmd: "docker logs -f --tail 100 ", Desc: "实时查看容器日志"},
	{Kind: suggestLinux, Cmd: "docker exec -it ", Desc: "在容器中执行交互命令"},
	{Kind: suggestLinux, Cmd: "docker restart ", Desc: "重启容器"},
	{Kind: suggestLinux, Cmd: "docker stop ", Desc: "停止容器"},
	{Kind: suggestLinux, Cmd: "docker start ", Desc: "启动容器"},
	{Kind: suggestLinux, Cmd: "docker rm -f ", Desc: "强制删除容器"},
	{Kind: suggestLinux, Cmd: "docker rmi ", Desc: "删除镜像"},
	{Kind: suggestLinux, Cmd: "docker pull ", Desc: "拉取远程镜像"},
	{Kind: suggestLinux, Cmd: "docker stats", Desc: "实时查看容器资源占用"},
	{Kind: suggestLinux, Cmd: "docker inspect ", Desc: "查看容器或镜像详情"},
	{Kind: suggestLinux, Cmd: "docker system df", Desc: "查看 Docker 磁盘占用"},
	{Kind: suggestLinux, Cmd: "docker system prune -af", Desc: "清理未使用的容器与镜像"},
	{Kind: suggestLinux, Cmd: "docker compose up -d", Desc: "后台启动 Compose 服务"},
	{Kind: suggestLinux, Cmd: "docker compose down", Desc: "停止并移除 Compose 容器"},
	{Kind: suggestLinux, Cmd: "docker compose logs -f", Desc: "实时查看 Compose 日志"},
	{Kind: suggestLinux, Cmd: "docker compose ps", Desc: "查看 Compose 服务状态"},
	{Kind: suggestLinux, Cmd: "docker compose restart", Desc: "重启 Compose 服务"},
	{Kind: suggestLinux, Cmd: "docker compose pull", Desc: "拉取 Compose 服务镜像"},

	// kubectl
	{Kind: suggestLinux, Cmd: "kubectl get pods", Desc: "查看当前命名空间 Pod"},
	{Kind: suggestLinux, Cmd: "kubectl get pods -A", Desc: "查看所有命名空间 Pod"},
	{Kind: suggestLinux, Cmd: "kubectl get svc", Desc: "查看 Service 列表"},
	{Kind: suggestLinux, Cmd: "kubectl get nodes -o wide", Desc: "查看节点列表及详情"},
	{Kind: suggestLinux, Cmd: "kubectl describe pod ", Desc: "查看 Pod 详情与事件"},
	{Kind: suggestLinux, Cmd: "kubectl logs -f ", Desc: "实时查看 Pod 日志"},
	{Kind: suggestLinux, Cmd: "kubectl exec -it ", Desc: "进入 Pod 容器终端"},
	{Kind: suggestLinux, Cmd: "kubectl apply -f ", Desc: "应用资源清单文件"},
	{Kind: suggestLinux, Cmd: "kubectl delete -f ", Desc: "删除资源清单定义的对象"},

	// git
	{Kind: suggestLinux, Cmd: "git status", Desc: "查看工作区状态"},
	{Kind: suggestLinux, Cmd: "git pull", Desc: "拉取并合并远程分支"},
	{Kind: suggestLinux, Cmd: "git push", Desc: "推送提交到远程仓库"},
	{Kind: suggestLinux, Cmd: "git checkout ", Desc: "切换分支或恢复文件"},
	{Kind: suggestLinux, Cmd: "git checkout -b ", Desc: "创建并切换新分支"},
	{Kind: suggestLinux, Cmd: "git branch -a", Desc: "查看所有本地与远程分支"},
	{Kind: suggestLinux, Cmd: "git log --oneline --graph -n 20", Desc: "查看简洁提交历史图"},
	{Kind: suggestLinux, Cmd: "git diff", Desc: "查看未暂存的文件改动"},
	{Kind: suggestLinux, Cmd: "git add .", Desc: "暂存所有改动"},
	{Kind: suggestLinux, Cmd: "git stash", Desc: "暂存当前工作区修改"},
	{Kind: suggestLinux, Cmd: "git stash pop", Desc: "恢复最近暂存的修改"},
	{Kind: suggestLinux, Cmd: "git fetch --all --prune", Desc: "同步所有远程分支引用"},

	// file / archive / permission / search
	{Kind: suggestLinux, Cmd: "ls -la", Desc: "列出全部文件及详细权限"},
	{Kind: suggestLinux, Cmd: "ls -lh", Desc: "以易读大小列出文件"},
	{Kind: suggestLinux, Cmd: "cd ..", Desc: "返回上级目录"},
	{Kind: suggestLinux, Cmd: "mkdir -p ", Desc: "递归创建目录"},
	{Kind: suggestLinux, Cmd: "rm -rf ", Desc: "递归强制删除文件或目录"},
	{Kind: suggestLinux, Cmd: "cp -r ", Desc: "递归复制目录"},
	{Kind: suggestLinux, Cmd: "mv ", Desc: "移动或重命名文件"},
	{Kind: suggestLinux, Cmd: "tar -zxvf ", Desc: "解压 .tar.gz 归档文件"},
	{Kind: suggestLinux, Cmd: "tar -zcvf ", Desc: "打包压缩为 .tar.gz"},
	{Kind: suggestLinux, Cmd: "tar -xvf ", Desc: "解压 tar 归档文件"},
	{Kind: suggestLinux, Cmd: "unzip ", Desc: "解压 zip 文件"},
	{Kind: suggestLinux, Cmd: "chmod +x ", Desc: "赋予可执行权限"},
	{Kind: suggestLinux, Cmd: "chmod -R 755 ", Desc: "递归设置 755 权限"},
	{Kind: suggestLinux, Cmd: "chown -R ", Desc: "递归修改文件所属用户与组"},
	{Kind: suggestLinux, Cmd: "find . -name ", Desc: "按名称查找当前目录下文件"},
	{Kind: suggestLinux, Cmd: "find / -name ", Desc: "全局按名称查找文件"},
	{Kind: suggestLinux, Cmd: "grep -rn ", Desc: "递归搜索文本内容并显示行号"},
	{Kind: suggestLinux, Cmd: "tail -f ", Desc: "实时跟踪文件末尾输出"},
	{Kind: suggestLinux, Cmd: "tail -n 100 ", Desc: "查看文件最后 100 行"},
	{Kind: suggestLinux, Cmd: "head -n 50 ", Desc: "查看文件前 50 行"},
	{Kind: suggestLinux, Cmd: "cat ", Desc: "查看文件内容"},
	{Kind: suggestLinux, Cmd: "less ", Desc: "分页查看文件内容"},

	// system / process / disk / network
	{Kind: suggestLinux, Cmd: "df -h", Desc: "查看磁盘挂载与剩余空间"},
	{Kind: suggestLinux, Cmd: "du -sh *", Desc: "查看当前目录各项占用大小"},
	{Kind: suggestLinux, Cmd: "du -h --max-depth=1", Desc: "查看一级子目录磁盘占用"},
	{Kind: suggestLinux, Cmd: "free -h", Desc: "查看内存与交换分区使用情况"},
	{Kind: suggestLinux, Cmd: "top", Desc: "实时查看系统进程与负载"},
	{Kind: suggestLinux, Cmd: "htop", Desc: "交互式进程查看器"},
	{Kind: suggestLinux, Cmd: "ps aux", Desc: "查看系统所有进程"},
	{Kind: suggestLinux, Cmd: "ps aux | grep ", Desc: "按关键字过滤进程"},
	{Kind: suggestLinux, Cmd: "kill -9 ", Desc: "强制终止指定 PID 进程"},
	{Kind: suggestLinux, Cmd: "pkill -f ", Desc: "按进程名匹配终止进程"},
	{Kind: suggestLinux, Cmd: "ss -tulpn", Desc: "查看监听端口及对应进程"},
	{Kind: suggestLinux, Cmd: "netstat -tulpn", Desc: "查看监听端口及对应进程"},
	{Kind: suggestLinux, Cmd: "lsof -i :", Desc: "查看指定端口占用情况"},
	{Kind: suggestLinux, Cmd: "ip addr", Desc: "查看网卡 IP 地址信息"},
	{Kind: suggestLinux, Cmd: "ip route", Desc: "查看路由表"},
	{Kind: suggestLinux, Cmd: "ping -c 4 ", Desc: "测试网络连通性"},
	{Kind: suggestLinux, Cmd: "curl -I ", Desc: "仅请求 HTTP 响应头"},
	{Kind: suggestLinux, Cmd: "curl -fsSL ", Desc: "静默请求并跟随重定向"},
	{Kind: suggestLinux, Cmd: "wget -c ", Desc: "断点续传下载文件"},
	{Kind: suggestLinux, Cmd: "scp -r ", Desc: "通过 SSH 递归传输文件"},
	{Kind: suggestLinux, Cmd: "rsync -avzP ", Desc: "增量同步文件并显示进度"},
	{Kind: suggestLinux, Cmd: "ssh-keygen -t ed25519", Desc: "生成 Ed25519 SSH 密钥对"},
	{Kind: suggestLinux, Cmd: "ufw status numbered", Desc: "查看防火墙状态与规则编号"},
	{Kind: suggestLinux, Cmd: "ufw allow ", Desc: "放行指定端口"},
	{Kind: suggestLinux, Cmd: "ufw enable", Desc: "启用 UFW 防火墙"},
	{Kind: suggestLinux, Cmd: "firewall-cmd --list-all", Desc: "查看 firewalld 防火墙规则"},
	{Kind: suggestLinux, Cmd: "firewall-cmd --reload", Desc: "重载 firewalld 防火墙"},
	{Kind: suggestLinux, Cmd: "nginx -t", Desc: "测试 Nginx 配置文件语法"},
	{Kind: suggestLinux, Cmd: "nginx -s reload", Desc: "平滑重载 Nginx 配置"},
	{Kind: suggestLinux, Cmd: "crontab -l", Desc: "查看当前用户定时任务"},
	{Kind: suggestLinux, Cmd: "crontab -e", Desc: "编辑当前用户定时任务"},
	{Kind: suggestLinux, Cmd: "tmux ls", Desc: "查看 tmux 会话列表"},
	{Kind: suggestLinux, Cmd: "tmux new -s ", Desc: "新建命名 tmux 会话"},
	{Kind: suggestLinux, Cmd: "tmux attach -t ", Desc: "连接指定 tmux 会话"},
	{Kind: suggestLinux, Cmd: "pm2 list", Desc: "查看 PM2 进程列表"},
	{Kind: suggestLinux, Cmd: "pm2 logs", Desc: "查看 PM2 实时日志"},
	{Kind: suggestLinux, Cmd: "pm2 restart all", Desc: "重启所有 PM2 进程"},
	{Kind: suggestLinux, Cmd: "uname -a", Desc: "查看内核与系统架构信息"},
	{Kind: suggestLinux, Cmd: "uptime", Desc: "查看系统运行时间与负载"},
	{Kind: suggestLinux, Cmd: "dmesg -T | tail -n 50", Desc: "查看最近内核日志"},
	{Kind: suggestLinux, Cmd: "history | tail -n 50", Desc: "查看最近执行的命令历史"},
}

// extractCursorPromptCommand extracts the command portion after a shell prompt
// delimiter such as "$ ", "# ", "% ", "❯ ", "➜ ", or "> ".
func extractCursorPromptCommand(line string) string {
	best := -1
	bestLen := 0
	for _, delim := range []string{"$ ", "# ", "% ", "❯ ", "➜ ", "> "} {
		if idx := strings.LastIndex(line, delim); idx >= 0 && idx+len(delim) > best+bestLen {
			best = idx
			bestLen = len(delim)
		}
	}
	if best < 0 {
		return ""
	}
	return line[best+bestLen:]
}

func (a *App) termSuggestions(rawInput string) []termSuggestion {
	q := strings.TrimLeft(rawInput, " \t")
	if q == "" {
		return nil
	}
	var snippets []termSuggestion
	if a.store != nil && a.unlocked() {
		for _, n := range a.store.Snippets() {
			body, err := a.snippetContent(n.ID)
			if err != nil {
				continue
			}
			body = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(body, "\r\n", "\n"), "\r", "\n"))
			if body == "" || strings.Contains(body, "\n") {
				continue
			}
			snippets = append(snippets, termSuggestion{
				Kind: suggestSnippet,
				Cmd:  body,
				Desc: n.Name,
			})
		}
	}
	return buildTermSuggestions(q, snippets)
}

func buildTermSuggestions(rawInput string, snippets []termSuggestion) []termSuggestion {
	q := strings.TrimLeft(rawInput, " \t")
	if q == "" {
		return nil
	}
	qLower := strings.ToLower(q)
	qTrim := strings.TrimSpace(qLower)
	if qTrim == "" {
		return nil
	}
	tokens := strings.Fields(qLower)

	var prefixMatches []termSuggestion
	var fuzzyMatches []termSuggestion
	seen := map[string]bool{}

	addCandidate := func(s termSuggestion, isPrefix bool) {
		key := strings.TrimSpace(s.Cmd)
		if key == "" || seen[key] {
			return
		}
		// Do not suggest a command identical to what is already typed unless
		// the candidate has a trailing space for further arguments and input does not.
		if strings.EqualFold(key, qTrim) && (!strings.HasSuffix(s.Cmd, " ") || strings.HasSuffix(q, " ")) {
			return
		}
		seen[key] = true
		if isPrefix {
			prefixMatches = append(prefixMatches, s)
		} else {
			fuzzyMatches = append(fuzzyMatches, s)
		}
	}

	matchTokens := func(targetLower string) bool {
		pos := 0
		for _, tk := range tokens {
			idx := strings.Index(targetLower[pos:], tk)
			if idx < 0 {
				return false
			}
			pos += idx + len(tk)
		}
		return true
	}

	// 1. Snippets (快捷指令)
	for _, s := range snippets {
		cmdLower := strings.ToLower(s.Cmd)
		nameLower := strings.ToLower(s.Desc)
		if strings.HasPrefix(cmdLower, qLower) || strings.HasPrefix(nameLower, qTrim) {
			addCandidate(s, true)
		} else if matchTokens(cmdLower) || strings.Contains(nameLower, qTrim) {
			addCandidate(s, false)
		}
	}

	// 2. Built-in Linux commands & subcommands (Linux基础命令及下接指令)
	sudoPrefix := ""
	linuxQuery := qLower
	if strings.HasPrefix(qLower, "sudo ") {
		rest := strings.TrimLeft(q[5:], " ")
		if rest != "" {
			sudoPrefix = q[:len(q)-len(rest)]
			linuxQuery = strings.ToLower(rest)
		}
	}
	linuxTokens := strings.Fields(linuxQuery)
	matchLinuxTokens := func(targetLower string) bool {
		if len(linuxTokens) == 0 {
			return false
		}
		// First token should match the base command prefix so unrelated commands don't flood
		fields := strings.Fields(targetLower)
		if len(fields) == 0 || !strings.HasPrefix(fields[0], linuxTokens[0]) {
			return false
		}
		pos := 0
		for _, tk := range linuxTokens {
			idx := strings.Index(targetLower[pos:], tk)
			if idx < 0 {
				return false
			}
			pos += idx + len(tk)
		}
		return true
	}

	for _, item := range linuxCommandCatalog {
		cand := item
		if sudoPrefix != "" {
			cand.Cmd = sudoPrefix + item.Cmd
		}
		cmdLower := strings.ToLower(item.Cmd)
		if strings.HasPrefix(cmdLower, linuxQuery) {
			addCandidate(cand, true)
		} else if matchLinuxTokens(cmdLower) {
			addCandidate(cand, false)
		}
	}

	out := append(prefixMatches, fuzzyMatches...)
	if len(out) > acMaxSuggestions {
		out = out[:acMaxSuggestions]
	}
	return out
}

func (v *termView) acOpen() bool {
	if v == nil || v.target == nil || v.acDismissed || v.offset != 0 || len(v.acItems) == 0 {
		return false
	}
	if v.target.Terminal().Modes().AltScreen {
		return false
	}
	return true
}

func (v *termView) resetAutocomplete() {
	v.inputBuf = ""
	v.acItems = nil
	v.acIndex = 0
	v.acScroll = 0
	v.acTouched = false
	v.acDismissed = false
}

func (v *termView) clampAcScroll() {
	if len(v.acItems) == 0 {
		v.acIndex, v.acScroll = 0, 0
		return
	}
	v.acIndex = max(0, min(v.acIndex, len(v.acItems)-1))
	if v.acIndex < v.acScroll {
		v.acScroll = v.acIndex
	} else if v.acIndex >= v.acScroll+acMaxVisible {
		v.acScroll = v.acIndex - acMaxVisible + 1
	}
	maxScroll := max(len(v.acItems)-acMaxVisible, 0)
	v.acScroll = max(0, min(v.acScroll, maxScroll))
}

func (v *termView) isPasswordPrompt(t *vt.Terminal) bool {
	if t == nil {
		return false
	}
	line := strings.ToLower(strings.TrimSpace(t.CursorLine()))
	if line == "" {
		return false
	}
	return strings.Contains(line, "password") ||
		strings.Contains(line, "passphrase") ||
		strings.Contains(line, "密码") ||
		strings.Contains(line, "verification code")
}

func (v *termView) cursorPrefix(t *vt.Terminal) string {
	if t == nil {
		return ""
	}
	snap := t.Snapshot(0)
	if snap.Modes.AltScreen || snap.CursorY < 0 || snap.CursorY >= len(snap.Lines) {
		return ""
	}
	cells := snap.Lines[snap.CursorY].Cells
	end := min(max(snap.CursorX, 0), len(cells))
	var sb strings.Builder
	for i := 0; i < end; i++ {
		c := cells[i]
		if c.Width == 0 {
			continue
		}
		if c.Rune == 0 {
			sb.WriteByte(' ')
		} else {
			sb.WriteString(c.Text())
		}
	}
	return sb.String()
}

func (v *termView) updateAutocomplete(t *vt.Terminal) {
	if t != nil && (t.Modes().AltScreen || v.isPasswordPrompt(t)) {
		v.acItems = nil
		v.acIndex, v.acScroll, v.acTouched = 0, 0, false
		return
	}
	q := strings.TrimLeft(v.inputBuf, " \t")
	if q == "" {
		v.acItems = nil
		v.acIndex, v.acScroll, v.acTouched = 0, 0, false
		return
	}
	if v.acProvider != nil {
		v.acItems = v.acProvider(v.inputBuf)
	} else {
		v.acItems = buildTermSuggestions(v.inputBuf, nil)
	}
	v.acIndex = 0
	v.acScroll = 0
	v.clampAcScroll()
}

func (v *termView) acceptAutocomplete(idx int) {
	if idx < 0 || idx >= len(v.acItems) {
		return
	}
	targetCmd := v.acItems[idx].Cmd
	cur := strings.TrimLeft(v.inputBuf, " \t")
	// Compute common byte prefix on rune boundaries
	common := 0
	for i := 0; i < len(cur) && i < len(targetCmd); {
		r1, w1 := utf8.DecodeRuneInString(cur[i:])
		r2, w2 := utf8.DecodeRuneInString(targetCmd[i:])
		if r1 != r2 || w1 != w2 {
			break
		}
		common += w1
		i += w1
	}
	delRunes := utf8.RuneCountInString(cur[common:])
	var payload string
	if delRunes > 0 {
		payload = strings.Repeat("\x7f", delRunes) + targetCmd[common:]
	} else {
		payload = targetCmd[common:]
	}
	if payload != "" {
		b := []byte(payload)
		v.send(b)
		v.fan(func(vt.Modes) []byte { return b })
	}
	v.inputBuf = targetCmd
	v.acTouched = false
	if v.target != nil {
		v.updateAutocomplete(v.target.Terminal())
	} else {
		v.acItems = nil
	}
}

func (v *termView) trackKey(key ui.Key, mods ui.Modifiers, t *vt.Terminal) {
	if t != nil && t.Modes().AltScreen {
		v.resetAutocomplete()
		return
	}
	switch {
	case mods == 0 && (key == ui.KeyEnter || key == ui.KeyTab):
		v.resetAutocomplete()
	case mods == 0 && key == ui.KeyBackspace:
		if v.inputBuf != "" {
			_, w := utf8.DecodeLastRuneInString(v.inputBuf)
			v.inputBuf = v.inputBuf[:len(v.inputBuf)-w]
			v.acTouched = false
			v.updateAutocomplete(t)
		} else {
			v.resetAutocomplete()
		}
	case (mods == ui.Ctrl && key == ui.KeyW) || (mods == ui.Alt && key == ui.KeyBackspace):
		s := strings.TrimRight(v.inputBuf, " \t")
		if idx := strings.LastIndexAny(s, " \t"); idx >= 0 {
			v.inputBuf = s[:idx+1]
		} else {
			v.inputBuf = ""
		}
		v.acTouched = false
		v.updateAutocomplete(t)
	case mods == ui.Ctrl && (key == ui.KeyU || key == ui.KeyC):
		v.resetAutocomplete()
	case mods == 0 && (key == ui.KeyUp || key == ui.KeyDown):
		v.resetAutocomplete()
	}
}

func (v *termView) trackText(text string, t *vt.Terminal) {
	if t != nil && (t.Modes().AltScreen || v.isPasswordPrompt(t)) {
		v.resetAutocomplete()
		return
	}
	if strings.ContainsAny(text, "\r\n") {
		norm := strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
		parts := strings.Split(norm, "\n")
		v.inputBuf = parts[len(parts)-1]
	} else {
		if v.inputBuf == "" {
			if seed := extractCursorPromptCommand(v.cursorPrefix(t)); seed != "" {
				v.inputBuf = seed
			}
		}
		v.inputBuf += text
	}
	v.acDismissed = false
	v.acTouched = false
	v.updateAutocomplete(t)
}

func suggestKindBadge(k termSuggestKind, tok theme.Tokens) (string, ui.Color) {
	switch k {
	case suggestSnippet:
		return "快捷", tok.Accent
	default:
		return "命令", tok.TextMuted
	}
}

// termAutocompleteView renders the floating suggestion list anchored to the
// terminal cursor inside termwrap.
func (a *App) termAutocompleteView(c *ui.Context, cn *conn, tv *termView, termBounds ui.Rect, focused bool, tok theme.Tokens) {
	if !focused || !tv.acOpen() {
		return
	}
	tv.metrics(1)
	caret := tv.caret()
	nItems := len(tv.acItems)
	vis := min(nItems, acMaxVisible)

	const (
		rowH   = float32(28)
		footH  = float32(22)
		padY   = float32(4)
		popupW = float32(360)
	)
	popupH := float32(vis)*rowH + footH + 2*padY + 2
	paneW, paneH := termBounds.W, termBounds.H
	if paneW <= 0 {
		paneW = 600
	}
	if paneH <= 0 {
		paneH = 400
	}
	w := min(popupW, max(paneW-16, 220))
	px := min(max(caret.X, 8), max(paneW-w-8, 8))
	py := caret.Y + caret.H + 4
	if py+popupH > paneH-8 && caret.Y-popupH-4 >= 4 {
		py = caret.Y - popupH - 4
	}

	card := ui.Column(c.Key("term:ac")).Absolute().Left(px).Top(py).Width(w).
		Padding(padY, 0).Radius(theme.RadiusControl).Clip().
		Background(tok.Inset).Border(1, tok.Line)
	theme.Shadow(tok.Lift).On(card)

	tv.clampAcScroll()
	end := min(tv.acScroll+vis, nItems)
	clickedIdx := -1
	card.Children(func() {
		for idx := tv.acScroll; idx < end; idx++ {
			i := idx
			item := tv.acItems[i]
			sel := tv.acTouched && i == tv.acIndex
			badgeText, badgeCol := suggestKindBadge(item.Kind, tok)

			row := ui.ButtonBase(c.Key("term:ac:item:"+strconv.Itoa(i))).
				Label(item.Cmd).FillWidth().Height(rowH).
				Padding(0, theme.Space(2.5)).Gap(theme.Space(2)).
				AlignItems(ui.Center).Justify(ui.Start).Cursor(ui.CursorPointer)
			switch {
			case sel:
				row.Background(tok.Accent.Alpha(accentWash(tok)))
			case row.Pressed():
				row.Background(tok.Press)
			case row.Hovered():
				row.Background(tok.Hover)
			}
			row.Children(func() {
				ui.Box(c).Padding(1, 5).Radius(theme.RadiusSmall).Shrink(0).
					Background(badgeCol.Alpha(0.15)).Children(func() {
					ui.Text(c, badgeText).FontSize(10).FontWeight(600).TextColor(badgeCol).SingleLine()
				})
				cmdCol := tok.Text
				if sel {
					cmdCol = tok.Accent
				}
				ui.Text(c, item.Cmd).Font(a.termFamily()).FontSize(theme.Footnote).FontWeight(500).
					TextColor(cmdCol).Grow(1).Shrink(1).SingleLine()
				if item.Desc != "" {
					ui.Text(c, item.Desc).FontSize(theme.Caption).TextColor(tok.TextMuted).
						Shrink(0).MaxWidth(120).SingleLine()
				}
			})
			if row.Clicked() {
				clickedIdx = i
			}
		}
		ui.Box(c).Height(1).Shrink(0).FillWidth().Background(tok.Line)
		ui.Row(c.Key("term:ac:foot")).FillWidth().Height(footH).Padding(0, theme.Space(2.5)).
			AlignItems(ui.Center).Children(func() {
			ui.Text(c, "↑↓ 切换 · Enter 补全 · Esc 关闭").FontSize(10).TextColor(tok.TextMuted).SingleLine()
			ui.Spacer(c)
			ui.Text(c, strconv.Itoa(tv.acIndex+1)+"/"+strconv.Itoa(nItems)).FontSize(10).TextColor(tok.TextMuted).SingleLine()
		})
	})
	if clickedIdx >= 0 {
		tv.acIndex = clickedIdx
		tv.acceptAutocomplete(clickedIdx)
		a.focusReq = cn.id
	}
}
