#!/usr/bin/env zsh

# تعريف الأنماط والألوان
autoload -U colors && colors
local bold=$'\e[1m'
local blue=$'\e[34m'
local green=$'\e[32m'
local red=$'\e[31m'
local yellow=$'\e[33m'
local magenta=$'\e[35m'
local cyan=$'\e[36m'
local reset=$'\e[0m'
local logo="${bold}${magenta}
   ▄▄▄▄▄▄▄ ▄▄▄▄▄▄▄ ▄▄▄▄▄▄▄ 
   █ ▄▄▄ █ ▄▄▄▄▄ █ ▄▄▄ █ 
   █ ███ █ █████ █ ███ █ 
   █▄▄▄▄▄█ █▄▄▄▄▄█ █▄▄▄▄▄█ 
   ▄▄▄▄▄ ▄▄▄ ▄▄▄▄▄ ▄▄▄ 
   ▀▀▀████▄▄ ██▄▄▄▀▀▀▀▀▀ 
     ▄▄█████▄▄▄▄    ▄▄▄ 
    █████▀▀▀███████████ 
   ✨ Majd Search Engine ✨
${reset}"

# دالة المساعدة المُحسنة
show_help() {
  print -P "%B${logo}%b"
  print -P "%B${cyan}Usage:%b ${green}searchit [options] <keywords>${reset}"
  print -P ""
  print -P "%B${yellow}Options:%b"
  print -P "  ${green}-h, --help        %fShow this help message"
  print -P "  ${green}-n, --number <N> %fNumber of results (default: 5)"
  print -P "  ${green}-s, --site <URL> %fSearch within specific site"
  print -P "  ${green}-t, --type <EXT> %fFilter by file type (pdf, docx...)"
  print -P "  ${green}-e, --engine <E> %fSearch engine (${magenta}google${reset}/bing/duckduckgo)"
  print -P "  ${green}-m, --mode <M>   %fOutput mode (${blue}summary${reset}/full/links)"
  print -P ""
  print -P "%B${yellow}Examples:%b"
  print -P "  ${cyan}searchit \"quantum computing\""
  print -P "  ${cyan}searchit -n 10 -t pdf \"AI research papers\""
  print -P "  ${cyan}searchit --site github.com \"neural networks\""
  print -P "  ${cyan}searchit -e bing -m links \"latest tech news\""
}

# معالجة الأخطاء المحسنة
handle_error() {
  print -P "%B${red}✖ Error:%b ${1}${reset}" >&2
  return 1
}

# دالة البحث الرئيسية
searchit() {
  local -a search_term
  local num_results=5
  local search_site file_type search_engine="google" output_mode="summary"
  local engine_arg site_arg type_arg mode_arg

  # معالجة الخيارات باستخدام zparseopts
  zmodload zsh/zutil
  zparseopts -E -D -A opts \
    h=help -help=help \
    n:=num_results -number:=num_results \
    s:=site_arg -site:=site_arg \
    t:=type_arg -type:=type_arg \
    e:=engine_arg -engine:=engine_arg \
    m:=mode_arg -mode:=mode_arg

  # عرض المساعدة إذا طُلب
  if [[ -n "$help" ]]; then
    show_help
    return 0
  fi

  # معالجة الوسائط
  search_term=("${@[@]}")

  # التحقق من وجود كلمات بحث
  if [[ ${#search_term} -eq 0 ]]; then
    read -r "?${bold}${cyan}➤ Search query: ${reset}" search_term
    [[ -z "$search_term" ]] && return 0
  fi

  # تعيين القيم من الخيارات
  search_engine=${${(L)opts[--engine]:-${opts[-e]}}:-google}
  output_mode=${${(L)opts[--mode]:-${opts[-m]}}:-summary}
  search_site=${opts[--site]:-${opts[-s]}}
  file_type=${opts[--type]:-${opts[-t]}}
  num_results=${opts[--number]:-${opts[-n]:-5}}

  # بناء استعلام البحث
  local -a query_parts=("${search_term[@]}")
  [[ -n $search_site ]] && query_parts+=("site:${search_site}")
  [[ -n $file_type ]] && query_parts+=("filetype:${file_type}")

  # بناء رابط البحث
  local search_url
  case $search_engine in
    google)
      search_url="https://www.google.com/search?num=${num_results}&q=${(j:+:)query_parts}"
      ;;
    bing)
      search_url="https://www.bing.com/search?count=${num_results}&q=${(j:+:)query_parts}"
      ;;
    duckduckgo)
      search_url="https://duckduckgo.com/?q=${(j:+:)query_parts}"
      ;;
    *)
      handle_error "Unsupported engine: ${search_engine}"
      return 1
      ;;
  esac

  # جلب وعرض النتائج
  print -P "%B${yellow}➤ Searching ${magenta}${search_engine}${yellow} for:%b ${green}${query_parts[@]}${reset}\n"

  local html_content=$(curl -s -A "Mozilla/5.0" "$search_url")

  case $output_mode in
    full|summary|links)
      # استخراج النتائج حسب المحرك
      case $search_engine in
        google)
          local results=(${(f)"$(echo $html_content | 
            pup 'h3, a[href^="/url?"] json{}' |
            jq -r '.[] | .text, .href' |
            sed -E 's/\/url\?q=//;s/&.*//')"})
          ;;
        bing)
          local results=(${(f)"$(echo $html_content |
            pup 'h2 a json{}' |
            jq -r '.[] | .text, .href')"})
          ;;
        duckduckgo)
          local results=(${(f)"$(echo $html_content |
            pup 'a.result__a json{}' |
            jq -r '.[] | .text, .href')"})
          ;;
      esac

      # عرض النتائج
      case $output_mode in
        full)
          for i in {1..${#results}}; do
            if (( i % 2 == 1 )); then
              print -P "%B${green}${results[i]}%b"
            else
              print -P "${blue}${results[i]}%b\n"
            fi
          done
          ;;
        links)
          print -P "%B${cyan}🔗 Direct links:%b\n"
          for link in ${results[@]}; do
            [[ $link =~ ^http ]] && print -P "${blue}${link}${reset}"
          done
          ;;
        *)
          print -P "%B${cyan}📚 Top results:%b\n"
          for ((i=1; i<=$#results; i+=2)); do
            print -P "%B${green}${results[i]}%b"
            print -P "${blue}${results[i+1]}%b\n"
          done
          ;;
      esac
      ;;
    *)
      handle_error "Invalid output mode: ${output_mode}"
      return 1
      ;;
  esac

  # فتح الرابط الأول تلقائيًا
  if [[ -n $results[2] ]]; then
    print -P "\n${yellow}🚀 Opening first result...${reset}"
    xdg-open "${results[2]}" &>/dev/null
  fi
}

# تعريف الأوامر والاختصارات
alias sh="noglob searchit"
alias srch="searchit"

# تحميل التبعيات
if ! (( $+commands[pup] )); then
  print -P "%B${red}✖ Required dependency 'pup' missing! Install with:"
  print -P "%B${cyan}  sudo apt install golang && go install github.com/ericchiang/pup@latest%b"
fi

if ! (( $+commands[jq] )); then
  print -P "%B${red}✖ Required dependency 'jq' missing! Install with:"
  print -P "%B${cyan}  sudo apt install jq%b"
fi