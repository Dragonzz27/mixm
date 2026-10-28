# SSH 登录和交互式 Bash 共用开发虚拟环境。
if [ -d /opt/venv ]; then
    export VIRTUAL_ENV=/opt/venv
    case ":${PATH}:" in
        *:/opt/venv/bin:*) ;;
        *) export PATH="/opt/venv/bin:${PATH}" ;;
    esac
fi
