package main;

import(
	"os"
	"fmt"
	"bytes"
	"strings"
	"io/ioutil"
	"crypto/sha1"
	"encoding/hex"
	"compress/zlib"
	"path/filepath"
)

func initializeRepo(name string){
	for _, dir := range []string{".mygit", ".mygit/objects"}{
		if err := os.MkdirAll(dir, 0755); err != nil{
			fmt.Printf("Error creating directory: %s\n", err);
			os.Exit(2);
		}
	}

	if err := ioutil.WriteFile(".mygit/LATEST_COMMIT", []byte(""), 0644); err != nil{
		fmt.Printf("Error writing file: %s\n", err);
		os.Exit(2);
	}

	config_name := fmt.Sprintf("name = %s", name);
	if err:= ioutil.WriteFile(".mygit/CONFIG", []byte(config_name), 0644); err != nil{
		fmt.Printf("Error writing file: %s\n", err);
		os.Exit(2);
	}

	fmt.Printf("Initialized empty mygit repository\n");
}

func generateHashAndCompress(contentsToHash []byte) string{
	hasher := sha1.New();
	hasher.Write([]byte(contentsToHash));
	hash := hasher.Sum(nil);
	sha := hex.EncodeToString(hash);
	dirName := sha[:2];
	fileName := sha[2:];
	dirPath := filepath.Join(".mygit", "objects", string(dirName));
	filePath := filepath.Join(dirPath, string(fileName));

	if err := os.MkdirAll(dirPath, 0755); err != nil{
		fmt.Printf("Error creating directory %s: %s\n", dirPath, err);
		os.Exit(3);
	}

	var buffer bytes.Buffer;
	writer := zlib.NewWriter(&buffer);
	writer.Write([]byte(contentsToHash));
	writer.Close();

	if err := ioutil.WriteFile(filePath, buffer.Bytes(), 0755); err != nil{
		fmt.Printf("Error writing file %s: %s\n", filePath, err);
		os.Exit(3);
	}

	return sha;
}

func hashObject(file string) string{
	fileData, err := ioutil.ReadFile(file);
	if(err != nil){
		fmt.Printf("Error reading file %s: %s\n", file, err);
		os.Exit(4);
	}

	return generateHashAndCompress(fileData);
}

func catFile(subdir string, file string){
	filePath := filepath.Join(".mygit", "objects", subdir, file);
	data, err := ioutil.ReadFile(filePath);
	if(err != nil){
		fmt.Printf("Error reading file %s: %s\n", filePath, err);
		os.Exit(5);
	}

	reader, err := zlib.NewReader(bytes.NewReader(data));
	if err != nil{
		fmt.Printf("Error decompressing blob for SHA %s: %s\n", os.Args[3], err);
		os.Exit(5);
	}

	decompressedData, err := ioutil.ReadAll(reader);
	if err != nil{
		fmt.Printf("Error reading blob for SHA %s: %s\n", os.Args[3], err);
		os.Exit(5);
	}

	reader.Close();

	fmt.Printf(string(decompressedData));
}

func writeTree(path string) string{
	treeEntries := []string{};

	directory, err := ioutil.ReadDir(path);
	if err != nil{
		fmt.Printf("Error reading directory: %s\n", err);
		os.Exit(6);
	}

	for _, item := range directory{
		if (item.Name() == ".mygit") || (item.Name() == ".git"){
			continue;
		}

		if item.IsDir(){
			// Something is wrong with trying to write trees of subdirectories.
			// I always get the "no such file or directory" error, even though the file clearly exists in the subdirectory
			treeSha := writeTree(filepath.Join(path, item.Name()));
			entry := fmt.Sprintf("%s: %s (tree)\n", item.Name(), treeSha);
			treeEntries = append(treeEntries, entry);
		}else{
			fileSha := hashObject(item.Name());
			entry := fmt.Sprintf("%s: %s (blob)\n", item.Name(), fileSha);
			treeEntries = append(treeEntries, entry);
		}
	}

	var treeData bytes.Buffer;
	for _, entry := range treeEntries{
		treeData.WriteString(entry);
	}

	return generateHashAndCompress(treeData.Bytes());
}

func commitTree(commitMessage string) string{
	commitEntries := []string{};

	treeSha := writeTree(".");
	entry := fmt.Sprintf("tree: %s\n", treeSha);
	commitEntries = append(commitEntries, entry);

	parentCommitSha, err := ioutil.ReadFile(".mygit/LATEST_COMMIT");
	if(err != nil){
		fmt.Printf("Error reading file .mygit/LATEST_COMMIT: %s\n", err);
		os.Exit(7);
	}else if(string(parentCommitSha) != ""){
		entry := fmt.Sprintf("parent: %s\n", parentCommitSha);
		commitEntries = append(commitEntries, entry);
	}

	configFileContents, err := ioutil.ReadFile(".mygit/CONFIG");
	if(err != nil){
		fmt.Printf("Error reading file .mygit/CONFIG: %s\n", err);
		os.Exit(7);
	}

	contentString := string(configFileContents);
	nameParts := strings.Split(contentString, "=");
	if len(nameParts) != 2{
		fmt.Printf("Invalid Config File Format: .mygit/CONFIG\n");
		os.Exit(7);
	}

	name := strings.TrimSpace(nameParts[1]);
	entry = fmt.Sprintf("committer: %s\n", name);
	commitEntries = append(commitEntries, entry);

	entry = fmt.Sprintf("\nmessage: %s\n", commitMessage);
	commitEntries = append(commitEntries, entry);

	var commitData bytes.Buffer;
	for _, entry := range commitEntries{
		commitData.WriteString(entry);
	}

	commitSha := generateHashAndCompress(commitData.Bytes());
	if err := ioutil.WriteFile(".mygit/LATEST_COMMIT", []byte(commitSha), 0755); err != nil{
		fmt.Printf("Error writing file .mygit/LATEST_COMMIT: %s\n", err);
		os.Exit(7);
	}

	return commitSha;
}

func printCommitHistory(commitSha string){
	dirName := commitSha[:2];
	fileName := commitSha[2:];
	commitFilePath := filepath.Join(".mygit", "objects", dirName, fileName);

	commitContent, err := ioutil.ReadFile(commitFilePath);
	if err != nil{
		fmt.Printf("Error reading commit file: ", err);
		os.Exit(8);
	}

	reader, err := zlib.NewReader(bytes.NewReader(commitContent));
	if err != nil{
		fmt.Printf("Error decompressing commit file: ", err);
		os.Exit(8);
	}

	decompressedData, err := ioutil.ReadAll(reader);
	if err != nil{
		fmt.Printf("Error reading blob for SHA %s: %s\n", os.Args[3], err);
		os.Exit(5);
	}

	reader.Close();

	commitLines := strings.Split(string(decompressedData), "\n");
	var parent, commitMessage string;
	for _, line := range commitLines{
		if strings.HasPrefix(line, "parent:"){
			parent = strings.TrimSpace(strings.TrimPrefix(line, "parent: "));
		}else if strings.HasPrefix(line, "message:"){
			commitMessage = strings.TrimSpace(strings.TrimPrefix(line, "message: "));
		}
	}

	fmt.Printf("* \x1b[33m%s\x1b[0m %s\n", commitSha, commitMessage);

	if parent != ""{
		printCommitHistory(parent);
	}
}

func printCommands(){
	fmt.Printf("./mygit.sh init -n <your_name> to initialize a mygit repository\n");
	fmt.Printf("./mygit.sh hash-object -w <file> to store the file in the database\n");
	fmt.Printf("./mygit.sh write-tree to store all files in the project in the database\n");
	fmt.Printf("./mygit.sh commit -m <message> to store a commit in the database\n");
	fmt.Printf("./mygit.sh cat-file -p <object SHA> to see the contents of the object associated with the SHA\n");
	fmt.Printf("./mygit.sh log to print your commit history\n");
}

func main(){
	if(len(os.Args) < 2){
		fmt.Printf("usage: ./mygit.sh <command> [args...]\nType \"./mygit.sh help\" for a list of available commands\n");
		return;
	}

	if(os.Args[1] == "help"){
		if(len(os.Args) > 2){
			fmt.Printf("ERROR: Too many arguments\nusage: ./mygit.sh help\n");
		}
		printCommands();
		return;
	}else if(os.Args[1] == "init"){
		if(len(os.Args) < 4){
			fmt.Printf("usage: ./mygit.sh init -n <your_name>\n");
			return;
		}else if(len(os.Args) > 4){
			fmt.Printf("usage: ./mygit.sh init -n <your_name>\nIf you have spaces in your name, please wrap the name in quotes.\n");
			return;
		}
	
		if(os.Args[2] != "-n"){
			fmt.Printf("unknown flag: %s\n", os.Args[2]);
			return;
		}

		initializeRepo(os.Args[3]);
	}else if(os.Args[1] == "hash-object"){
		if(len(os.Args) < 4){
			fmt.Printf("usage: ./mygit.sh hash-object -w <file>\n");
			return;
		}else if(len(os.Args) > 4){
			fmt.Printf("usage: ./mygit.sh init -n <file>\nIf the file or path to the file contains spaces, please wrap the path in quotes.\n");
			return;
		}
	
		if(os.Args[2] != "-w"){
			fmt.Printf("unknown flag: %s\n", os.Args[2]);
			return;
		}

		sha := hashObject(os.Args[3]);
		fmt.Printf("%s\n", sha);
	}else if(os.Args[1] == "cat-file"){
		if(len(os.Args) < 4){
			fmt.Printf("usage: ./mygit.sh cat-file -p <object SHA>\n");
			return;
		}else if(len(os.Args) > 4){
			fmt.Printf("ERROR: Too many arguments\nusage: ./mygit.sh cat-file -p <object SHA>\n");
			return;
		}
	
		if(os.Args[2] != "-p"){
			fmt.Printf("unknown flag: %s\n", os.Args[2]);
			return;
		}

		catFile(os.Args[3][:2], os.Args[3][2:]);
	}else if(os.Args[1] == "write-tree"){
		if(len(os.Args) > 2){
			fmt.Printf("Error: Too many arguments\nusage: ./mygit.sh write-tree");
			return;
		}
		sha := writeTree(".");
		fmt.Printf("%s\n", sha);
	}else if(os.Args[1] == "commit"){
		if(len(os.Args) < 4){
			fmt.Printf("usage: ./mygit.sh commit -m <message>");
		}else if(len(os.Args) > 4){
			fmt.Printf("usage: ./mygit.sh commit -m <message>\nIf your commit message contains spaces, please wrap the message in quotes.\n");
			return;
		}

		if(os.Args[2] != "-m"){
			fmt.Printf("unknown flag: %s\n", os.Args[2]);
			return;
		}

		sha := commitTree(os.Args[3]);
		fmt.Printf("%s\n", sha);
	}else if(os.Args[1] == "log"){
		if(len(os.Args) > 2){
			fmt.Printf("usage: ./mygit.sh log\n");
			return;
		}

		latestCommitSha, err := ioutil.ReadFile(".mygit/LATEST_COMMIT");
		if err != nil{
			fmt.Printf("Error reading file LATEST_COMMIT: ", err);
			return;
		}

		printCommitHistory(string(latestCommitSha));
	}else{
		fmt.Printf("Unknown command %s\n", os.Args[1]);
		return;
	}
}