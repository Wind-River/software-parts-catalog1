<?php

require 'key.php';

function Encrypt(?string $Content, string $Key): string {
    $IV = random_bytes(12);
    return $IV . openssl_encrypt($Content, 'aes-256-gcm', $Key, OPENSSL_RAW_DATA, $IV, $Tag, '', 16) . $Tag;
}

$key = base64url_decode($argv[1]);
echo bin2hex(Encrypt($argv[2], $key));