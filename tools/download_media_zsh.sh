#!/bin/bash

download_media() {
    # تعريف الألوان
    RED='\033[1;31m'
    GREEN='\033[1;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[1;34m'
    MAGENTA='\033[1;35m'
    CYAN='\033[1;36m'
    BG_RED='\033[48;5;196m'
    BG_GREEN='\033[48;5;40m'
    NC='\033[0m'

    # دالة تثبيت التبعيات
    _install_deps() {
        echo -e "\n${YELLOW}▶ جاري تثبيت المتطلبات...${NC}"

        if ! sudo -v; then
            echo -e "${RED}✖ خطأ في الصلاحيات!${NC}"
            return 1
        fi

        echo -e "${CYAN}⏳ تحديث قوائم الحزم...${NC}"
        sudo apt-get update -qq

        echo -e "${CYAN}⏳ تثبيت الحزم الأساسية...${NC}"
        sudo apt-get install -y -qq \
            yt-dlp \
            ffmpeg \
            atomicparsley \
            python3 \
            python3-pip \
            libavcodec-extra \
            imagemagick \
            jq

        echo -e "${CYAN}⏳ تحديث مكتبات بايثون...${NC}"
        python3 -m pip install -q --upgrade mutagen yt-dlp

        echo -e "${GREEN}✅ اكتمل التثبيت!${NC}"
    }

    # واجهة البداية
    echo -e "${MAGENTA}"
    cat << "EOF"

      ___      _____ _   _  ______ _   _ _______ 
     / _ \    |  __ \ | | | |  ___| \ | |_   _|
    / /_\ \   | |  \/ | | | | |_  |  \| | | |  
    |  _  |   | | __| | | | |  _| | . ` | | |  
    | | | |   | |_\ \ |_| | | |___| |\  | | |  
    \_| |_/    \____/\___/  \____/\_| \_/ \_/  
                                                
    Crafted for Majd by DeepSeek
    Enhanced with finesse by Gemini
EOF
    echo -e "${NC}"

    # التحقق من التبعيات
    if ! command -v yt-dlp &>/dev/null || ! command -v ffmpeg &>/dev/null || ! command -v jq &>/dev/null; then
        echo -e "${BG_RED}⚠️  التبعيات غير مثبتة!${NC}"
        echo -ne "هل تريد التثبيت التلقائي؟ [Y/n] "
        read choice
        if [[ "${choice:-Y}" =~ [yY] ]]; then
            _install_deps || return 1
        else
            echo -e "${RED}✖ تم الإلغاء${NC}"
            return 1
        fi
    fi

    # مدخلات المستخدم
    while true; do
        echo -e "${CYAN}▶ أدخل رابط الفيديو (أو 'exit' للخروج):${NC}"
        echo -ne "> "
        read video_url
        case "$video_url" in
            exit|exit) return 0 ;;
            "")
                echo -e "${RED}✖ الرابط لا يمكن أن يكون فارغًا!${NC}"
                continue
                ;;
            *)
                if [[ "$video_url" =~ ^https?:// ]]; then
                    break
                else
                    echo -e "${RED}✖ رابط غير صالح!${NC}"
                fi
                ;;
        esac
    done

    # اختيار التنسيق
    PS3="$(echo -e "${YELLOW}? اختر نوع الملف: ${NC}")"
    select format in "فيديو (MP4)" "صوت (MP3)" "إلغاء"; do
        case $format in
            "فيديو (MP4)")
                file_ext="mp4"
                format_str="bestvideo[ext=mp4]+bestaudio[ext=m4a]/best"

                # Quality selection (only if MP4 is chosen)
                echo -e "${YELLOW}? اختر جودة التنزيل:${NC}"
                select quality in "أعلى جودة" "4K" "1080p" "720p" "480p"; do
                    case $quality in
                        "أعلى جودة") q_flag="" ;;
                        "4K") q_flag="[height<=2160]" ;;
                        "1080p") q_flag="[height<=1080]" ;;
                        "720p") q_flag="[height<=720]" ;;
                        "480p") q_flag="[height<=480]" ;;
                        *) echo -e "${RED}✖ اختيار غير صالح!${NC}" ;;
                    esac
                    break
                done

                break
                ;;
            "صوت (MP3)")
                file_ext="mp3"
                format_str="bestaudio/best"
                q_flag="" # Reset q_flag for MP3
                break
                ;;
            "إلغاء")
                return 0
                ;;
            *)
                echo -e "${RED}✖ اختيار غير صالح!${NC}"
                ;;
        esac
    done

    # بناء الأمر
    cmd="yt-dlp --no-playlist --ignore-errors"
    cmd+=" -f '${format_str}${q_flag}'"
    cmd+=" -o '%(title)s.%(ext)s'"

    # إضافة خصائص خاصة للصوت
    if [[ "$file_ext" == "mp3" ]]; then
        cmd+=" --extract-audio --audio-format mp3"
        cmd+=" --embed-thumbnail --ppa 'EmbedThumbnail+ffmpeg_o:-c:v mjpeg -vf crop=ih*9/16:ih'"
    fi

    # التنفيذ
    echo -e "\n${BLUE}▶ جاري التنزيل...${NC}"
    eval "${cmd} '$video_url' 2> >(grep -v 'WARNING:')"

    # النتيجة
    if [[ $? -eq 0 ]]; then
        echo -e "${BG_GREEN}✅ تم التنزيل بنجاح!${NC}"
        # إعادة الحصول على اسم الملف الصحيح بعد التنزيل
        local downloaded_file=$(ls -t ./*.${file_ext} | head -1)
        echo -e "${GREEN}→ الموقع: $(pwd)/$downloaded_file${NC}"
    else
        echo -e "${BG_RED}✖ فشل التنزيل!${NC}"
        echo -e "${RED}الأسباب المحتملة:"
        echo -e "- الرابط محمي أو غير متاح"
        echo -e "- تنسيق غير مدعوم"
        echo -e "- مشكلة في اتصال الإنترنت"
        echo -e "- النظام الأساسي غير مدعوم${NC}"
    fi
}

# اختصار إضافي مع دعم Zsh/Bash
if [[ -n "$ZSH_VERSION" ]]; then
    alias dm="download_media"
elif [[ -n "$BASH_VERSION" ]]; then
    alias dm=download_media
fi