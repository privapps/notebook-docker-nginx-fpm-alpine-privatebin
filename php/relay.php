<?php

define('REMOTE_SUPPORT', true);


class Targets {
	private $servers = [
		"https://paste.eccologic.net",
		"https://paste.i2pd.xyz/",
		"https://pastebin.hot-chilli.net/",
		"https://pb.florian2833z.de/",
		"https://bin.moritz-fromm.de/",
		"https://paste.fizi.ca/",
		"https://pastebin.grey.pw/",
		"https://paste.tuxcloud.net/",
		"https://paste.taiga-san.net/",
		'https://vim.cx/',
		'https://privatebin.at/',
		'https://zerobin.farcy.me/',
		'https://snip.dssr.ch/',
		'https://bin.snopyta.org/',
		'https://paste.danielgorbe.com/',
		'https://pastebin.aquilenet.fr/',
		'https://pb.nwsec.de/',
		'https://wtf.roflcopter.fr/paste/',
		'https://paste.systemli.org/',
		'https://bin.acquia.com/'
	];

	function getServers(){
		return $this->servers;
	}
}

class Worker{
	private $servers;
	
	function __construct() {
		$this->servers = (new Targets())->getServers();
	}

	function getAServer(){
		$p = mt_rand(0,count($this->servers)-1);
		return $this->servers[$p];
	}
	function relay( $data_string, $count=3){
		$sv = $this->getAServer();
		$slash = strpos ($sv, "/" , 10 );
		$prefix = $slash ? substr($sv, 0, $slash) : $sv;
		$ch = curl_init($sv);
		$options = array(
			CURLOPT_RETURNTRANSFER => true,     // return web page
			CURLOPT_HEADER         => false,    // don't return headers
			CURLOPT_FOLLOWLOCATION => true,     // follow redirects
			CURLOPT_ENCODING       => "",       // handle all encodings
			CURLOPT_AUTOREFERER    => true,     // set referer on redirect
			CURLOPT_CONNECTTIMEOUT => 5,      // timeout on connect
			CURLOPT_TIMEOUT        => 15,      // timeout on response
			CURLOPT_MAXREDIRS      => 10,       // stop after 10 redirects
			CURLOPT_SSL_VERIFYPEER => false,     // Disabled SSL Cert checks
			CURLOPT_POSTFIELDS => $data_string,
			CURLOPT_POST => 1,
			CURLOPT_HTTPHEADER =>  array(
				'Content-Type: application/json',
				'Content-Length: ' . strlen($data_string),
				'X-Requested-With: JSONHttpRequest',
				'origin: '. $prefix
			)
		);
		curl_setopt_array( $ch, $options );
		$result = curl_exec($ch);
		curl_close($ch);
		$json = json_decode($result, true);
		header('Content-Type: application/json');
		if(isset($json['status']) && $json['status'] == 0 ){
			$json['url'] = $prefix.$json['url'];
			echo json_encode($json);
		} else if($count > 0){ // retry
			$this->relay($data_string, $count-1);
		} else {
			echo $result;
		}
	}

}

$x = new Worker();

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
	$foo = file_get_contents("php://input");
	header('Content-Type: application/json');
	$json = isset($foo) ? json_decode($foo, true) : '';


	if(!isset($json) || !isset($json['v']) || !isset($json['ct']) || !isset($json['adata'])) {
		http_response_code(404);
		die();
	}
	if( REMOTE_SUPPORT ) {
	header('Access-Control-Allow-Origin: *');
	header("Access-Control-Allow-Methods: POST");
	header('Access-Control-Allow-Headers: Content-Type, X-Requested-With');
	}
	$x->relay($foo);
	http_response_code(201);
} if ($_SERVER['REQUEST_METHOD'] === 'OPTIONS'){
	if( REMOTE_SUPPORT ) {
        header('Access-Control-Allow-Origin: *');
        header("Access-Control-Allow-Methods: POST");
        header('Access-Control-Allow-Headers: Content-Type, X-Requested-With');
        }
} else{
	http_response_code(200);
}
