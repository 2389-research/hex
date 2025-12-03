#!/bin/bash
echo "Testing theme display:"
echo ""
echo "If your terminal supports colors, you should see:"
echo "1. Three different colored blocks below"
echo "2. Gradients in each line"
echo ""

# Dracula gradient (purple to pink)
echo -n "Dracula: "
for i in {0..20}; do
    r=$((189 + (255-189)*i/20))
    g=$((147 - (147-121)*i/20))
    b=$((249 - (249-198)*i/20))
    printf "\033[38;2;%d;%d;%dm█\033[0m" $r $g $b
done
echo ""

# Gruvbox gradient (orange to red-orange)
echo -n "Gruvbox: "
for i in {0..20}; do
    r=$((254 - (254-214)*i/20))
    g=$((128 - (128-93)*i/20))
    b=$((25 - (25-14)*i/20))
    printf "\033[38;2;%d;%d;%dm█\033[0m" $r $g $b
done
echo ""

# Nord gradient (cyan to blue)
echo -n "Nord:    "
for i in {0..20}; do
    r=$((136 - (136-129)*i/20))
    g=$((192 - (192-161)*i/20))
    b=$((208 - (208-193)*i/20))
    printf "\033[38;2;%d;%d;%dm█\033[0m" $r $g $b
done
echo ""
echo ""
echo "If you DON'T see colors or gradients above, your terminal"
echo "may not support 24-bit color (truecolor)."
