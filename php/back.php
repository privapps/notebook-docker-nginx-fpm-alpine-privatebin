<?php

define('EMAIL_SUPPORTED', 0);
define('REMOTE_SUPPORT', 0);
define('DATA_ROOT', '/srv/data/');

define('EMAIL_TEMPLATE', <<<EMAIL
Please use following URL to update your private notes
%prefix%id=%id%&server=%key%&type=ed&new
You will obtain a symmetric key and choose your own end to end encryption key when you publish your first note. Please keep that full URL in private as no one but you knows it.
Please do not reply this email as it is not monitored.
EMAIL
);
if( REMOTE_SUPPORT ) {
	header('Access-Control-Allow-Origin: *');
	header("Access-Control-Allow-Methods: GET,POST,PUT,OPTIONS");
	header('Access-Control-Allow-Headers: Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With, X-Hash, X-Timestamp');
}
class Back{
	private $conf_file_loc = DATA_ROOT;
	private $data_file_loc;
	private $_db;
	function __construct() {
		$this->data_file_loc = $this->conf_file_loc . 'data/';
	}
	
	function register($json){
		$uq = uniqid();
		while (file_exists ($this->conf_file_loc.substr(uniqid(),-6))){
			$uq = uniqid();
		}
		$id = substr($uq,-6);
		$key = substr(hash('md5', uniqid()),-5);
		$file = fopen($this->conf_file_loc.$id, "w") or die("Unable to open file!");
		fwrite($file, $key);
		fclose($file);

		$object = (object) ['id' => $id, 'key' => $key];
		if(EMAIL_SUPPORTED){
			$data = array(
				'%prefix%' => $json['prefix'],
				'%id%' => $id,
				'%key%'=> $key
			);
			$strings = str_replace(array_keys($data), array_values($data), EMAIL_TEMPLATE);
			mail(
				$json['email'],
				'Your private notebook is ready',
				$strings,
				'From: Privapps Notebook'
			);
		}
		return $object;
	}
	
	function create_or_update($id, $data, $hash, $timestamp){
		if(!file_exists($this->conf_file_loc.$id)){
			return 404;
		}
		$key = file_get_contents($this->conf_file_loc.$id);
		if($hash !== hash('sha256', $data.$key)){
			return 401;
		}
		if(file_exists($this->data_file_loc.$id)){
			$ts = filemtime($this->data_file_loc.$id);
			if($ts != intval($timestamp)){
				return 409;
			}
		}
		$file = fopen($this->data_file_loc.$id, "w") or die('fail to open file');
		fwrite($file, $data);
		fclose($file);
		clearstatcache();
		$ts = filemtime($this->data_file_loc.$id);
		return array(201,$ts);
	}
	function get($id){
		if(!file_exists($this->data_file_loc.$id)){
			return false;
		}
		$data = file_get_contents($this->data_file_loc.$id);
		$ts = filemtime($this->data_file_loc.$id);
		return array($data, $ts);
	}
}

if (!function_exists('getallheaders')) {
    function getallheaders() {
    $headers = [];
    foreach ($_SERVER as $name => $value) {
        if (substr($name, 0, 5) == 'HTTP_') {
            $headers[str_replace(' ', '-', ucwords(strtolower(str_replace('_', ' ', substr($name, 5)))))] = $value;
        }
    }
    return $headers;
    }
}
$x = new Back();

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
	$foo = file_get_contents("php://input");
	header('Content-Type: application/json');
	$json = isset($foo) ? json_decode($foo,true) : '';
	if(EMAIL_SUPPORTED){
		if(!isset($json['email']) || !isset($json['prefix'])) {
			http_response_code(400);
			die();
		}
	}
	$data = $x->register($json);
	if(! EMAIL_SUPPORTED ){ // only send info if email is not supported.
		echo json_encode($data);
	}else{
		http_response_code(201);
	}
} else if ($_SERVER['REQUEST_METHOD'] === 'PUT') {
	$data = file_get_contents("php://input");
	$headers = getallheaders();
	$hash = $headers['X-Hash'];
	//var_dump($headers);
	$timestamp = isset($headers['X-Timestamp']) ? $headers['X-Timestamp'] : '0';
	$id = $_GET['pasteid'];
	if(!isset($_SERVER['QUERY_STRING']) || !isset($id) || !isset($hash) || !isset($data)){
		http_response_code(400);
		die();
	}
	$code = $x->create_or_update($_GET['pasteid'], $data, $hash, $timestamp);
	if(is_array($code)){
		echo json_encode([
			'status' => 0,
			'id' => $id,
			'url' => $id
		]);
		header('Content-Type: application/json');
		header('X-Timestamp: '.$code[1]);
		http_response_code(201);
	} else {
		http_response_code($code);
	}
} elseif($_SERVER['REQUEST_METHOD'] === 'OPTIONS'){
	// do sth
} elseif($_SERVER['REQUEST_METHOD'] === 'GET'){
	if(!isset($_SERVER['QUERY_STRING']) || !isset($_GET['pasteid'])){
		http_response_code(400);
		die();
	}
	$data = $x->get($_GET['pasteid']);
	if($data){
		header('Content-Type: application/json');
		header('X-Timestamp: '.$data[1]);
		http_response_code(200);
		echo $data[0];
	} else {
		http_response_code(404);
	}
} else {
	header($_SERVER["SERVER_PROTOCOL"]." 404 Not Found", true, 404);
}
