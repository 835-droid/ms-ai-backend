#!/bin/bash

# ~/.local/share/msai/remove_zsh.sh
# دالة "remove" متوافقة مع Oh My Zsh
# لتحميلها: أضف في ~/.zshrc السطر:
#    source ~/.local/share/msai/remove_zsh.sh

remove() {
  # تفعيل الألوان
  local RED='\033[0;31m' GREEN='\033[0;32m' YELLOW='\033[0;33m' BLUE='\033[0;34m' NC='\033[0m'
  local auto_confirm=0

  # مساعدة
  if [[ "$1" == "-h" || "$1" == "--help" ]]; then
    echo -e "${YELLOW}📖 استخدام: remove [خيارات] [اسم البرنامج]${NC}"
    echo "هذه الدالة تحذف البرنامج عبر APT، Snap، Flatpak، وماسح ملفات يدويًا."
    echo -e "${YELLOW}  خيارات:${NC}"
    echo "    -h, --help: عرض هذه المساعدة"
    echo "    -v: وضع الإسهاب (عرض المزيد من التفاصيل)"
    echo "    -y: الحذف التلقائي بدون تأكيد"
    return 0
  fi

  # معالجة الخيارات
  while [[ $# -gt 0 ]]; do
    case "$1" in
      -v) verbose=1; shift ;;
      -y) auto_confirm=1; shift ;;
      *) break ;;
    esac
  done

  # تأكد من وجود وسيط
  if [[ -z "$1" ]]; then
    echo -e "${RED}✖ يرجى تحديد اسم البرنامج.${NC}"
    return 1
  fi

  local pkg="$1"
  local pkg_lower=$(echo "$pkg" | tr '[:upper:]' '[:lower:]')
  local found=0

  echo -e "\n${BLUE}▐▓▒░ حذف شامل للبرنامج [$pkg]...${NC}\n"

  remove_apt "$pkg"
  remove_snap "$pkg"
  remove_flatpak "$pkg"
  remove_manual "$pkg"
  remove_config "$pkg"
  remove_system "$pkg"
  find_leftovers "$pkg"

  # انتهاء
  if [[ $found -gt 0 ]]; then
    echo -e "\n${GREEN}✓ تم حذف كل آثار $pkg بنجاح!${NC}"
    notify-send "🧹 Remove" "تم حذف $pkg"
    echo -e "\a"
  else
    echo -e "\n${RED}✖ لم يُعثر على $pkg${NC}"
    echo -e "\a"
  fi

  # سجل
  echo "$(date '+%d-%m-%Y %H:%M'): حذف $pkg" >> ~/.remove_log.txt
  tail -n100 ~/.remove_log.txt > ~/.remove_log.tmp && mv ~/.remove_log.tmp ~/.remove_log.txt
}

# الدوال المساعدة مع التعديلات
remove_flatpak() {
  local pkg="$1"
  if command -v flatpak &>/dev/null && flatpak list --app | grep -qi "$pkg"; then
    echo -e "${GREEN}├─ [FLATPAK]${NC}"
    flatpak uninstall --delete-data -y "$pkg" && flatpak uninstall --unused -y
    local result=$?
    if [[ $result -ne 0 ]]; then
      echo -e "${RED}│   └── فشل حذف Flatpak $pkg (رمز الخطأ: $result)${NC}"
    else
      echo -e "${GREEN}│   └── تم حذف Flatpak $pkg بنجاح${NC}"
      ((found++))
    fi
  fi
}

remove_manual() {
  local pkg="$1"
  local manual=("/usr/local/bin/$pkg" "/opt/$pkg" "$HOME/.local/bin/$pkg" "/usr/bin/$pkg" "/usr/lib/$pkg" "/usr/share/$pkg" "$HOME/msai/Downloads/$pkg")
  for p in "${manual[@]}"; do
    [[ -e "$p" ]] && {
      if [[ $auto_confirm -eq 1 ]]; then
        echo -e "${GREEN}├─ [MANUAL]${NC} $p"
        sudo rm -rfv "$p"
        ((found++))
      else
        read -p "${YELLOW}هل أنت متأكد من حذف ${p}؟ (y/n): ${NC}" confirm
        [[ "$confirm" == "y" ]] && {
          sudo rm -rfv "$p"
          ((found++))
        }
      fi
    }
  done
}

find_leftovers() {
  local pkg="$1"
  echo -e "${GREEN}├─ [SEARCH]${NC} (بحث شامل في النظام)"
  local exclude_dirs=("/proc" "/sys" "/dev" "/run" "/tmp")
  local find_opts="-regextype posix-extended -iregex .*${pkg}.* -type f -maxdepth 6 -print0"

  sudo find / \
    -xdev \
    -not \( -path "${exclude_dirs[0]}" -prune \) \
    -not \( -path "${exclude_dirs[1]}" -prune \) \
    -not \( -path "${exclude_dirs[2]}" -prune \) \
    -not \( -path "${exclude_dirs[3]}" -prune \) \
    -not \( -path "${exclude_dirs[4]}" -prune \) \
    $find_opts 2>/dev/null | while IFS= read -r -d $'\0' f; do
      echo -e "${YELLOW}│   ├─ وجد:${NC} $f"
      if [[ $auto_confirm -eq 1 ]]; then
        sudo rm -fv "$f" && ((found++))
      else
        read -p "${YELLOW}│   └── حذف؟ (y/n): ${NC}" confirm
        [[ "$confirm" == "y" ]] && sudo rm -fv "$f" && ((found++))
      fi
    done
}