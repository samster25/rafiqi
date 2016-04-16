if [ "$#" -lt 1 ]
then
    echo "Usage $0 path/to/data.csv"
    exit 1
fi

column -s, -t < $1 | less -#2 -N -S
