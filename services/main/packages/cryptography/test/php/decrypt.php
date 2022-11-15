<?php

require 'key.php';

function Decrypt(?string $Ciphertext, string $key): ?string {
    if (strlen($Ciphertext) < 28)
        return null;

    $IV = substr($Ciphertext, 0, 12);
    $Content = substr($Ciphertext, 12, -16);
    $Tag = substr($Ciphertext, -16, 16);

    try {
        return openssl_decrypt($Content, 'aes-256-gcm', $key, OPENSSL_RAW_DATA, $IV, $Tag);
    } catch (Exception $e) {
        return null;
    }
}

$key = base64url_decode($argv[1]);
echo Decrypt(hex2bin($argv[2]), $key);