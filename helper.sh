# Create a simple gem structure
mkdir my_test_gem
cd my_test_gem

# Create basic gem files
cat > my_test_gem.gemspec << 'EOF'
Gem::Specification.new do |spec|
  spec.name          = "my_test_gem"
  spec.version       = "1.0.0"
  spec.authors       = ["Your Name"]
  spec.email         = ["your.email@example.com"]
  spec.summary       = "A test gem"
  spec.description   = "A simple test gem for JFrog CLI"
  spec.files         = ["lib/my_test_gem.rb"]
  spec.require_paths = ["lib"]
end
EOF

mkdir lib
echo 'puts "Hello from my test gem!"' > lib/my_test_gem.rb

# Build the gem
gem build my_test_gem.gemspec

# Now you can publish it
cd ..
./jf pkg ruby publish my_test_gem/my_test_gem-1.0.0.gem --build-name=test --build-number=1
