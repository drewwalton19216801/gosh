#!/usr/bin/env gosh
# Test redirection functionality

echo "Testing output redirection..."
echo "This goes to a file" > output.txt
echo "This gets appended" >> output.txt

echo "Reading back the file:"
cat output.txt

echo "Testing input redirection:"
cat < input.txt

echo "Redirection test completed!"