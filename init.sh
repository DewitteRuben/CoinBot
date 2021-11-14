cat .bots | \
while read line; do
    token=$(echo $line | cut -d ' ' -f 1)
    coin=$(echo $line | cut -d ' ' -f 2)

    cmd="./bot -token $token -coin $coin"

    $cmd &

    sleep 1
done

sleep infinity