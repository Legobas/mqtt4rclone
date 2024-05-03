#!/bin/bash
usermod -o -u "$PUID" -g "$PGID" appuser
su -s /bin/bash appuser -c \
'/usr/bin/multirun "/usr/bin/rclone rcd --log-format none --log-level '"$RCLONE_LOGLEVEL"' --rc-web-gui-no-open-browser --rc-user cmd --rc-pass f6Lhi09wfbxkd8Ok2l4H" "/bin/app"'
