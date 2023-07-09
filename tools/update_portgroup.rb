#!/usr/bin/env ruby
#
# Update portGroup
#

ProjectROOT = File.join(__dir__, '..')
ProjectReadme = File.join(ProjectROOT, 'README.md')
MX1014GOFILE = File.join(ProjectROOT, 'mx1014/mx1014.go')
m = File.binread(ProjectReadme).match(/## Port Group\n```ruby([^`]+)/m)
abort '[!] portGroup not found' if m.nil?
portgroup = eval(m[1]).transform_values{ _1.split(',') }
3.times do 
  portgroup.each  { |key, ports|
    portgroup[key] = ports.map{|port| portgroup[port.to_sym] or port }.flatten(1)
  }
end
portgroup.transform_values!{|ports|
  res_ports = []
  ports.each do |port|
    if port.include? '-'
      sport, eport = port.split('-').map(&:to_i)
      res_ports += sport.upto(eport).to_a
    else
      res_ports << port.to_i
    end
  end
  res_ports.sort.uniq
}

portgroup_golang = portgroup.map{|name, ports| %Q|      "#{name}": []int{ #{ports.join(",")} },| }.join("\n")

gocode = File.binread(MX1014GOFILE)
if gocode.sub!(/(?<=portGroup = map\[string\]\[\]int \{\n)(.+?)(?=\n    \}\n)/m, portgroup_golang)
  File.binwrite(MX1014GOFILE, gocode)
else
  abort '[!] portGroup not found in mx1014.go'
end
