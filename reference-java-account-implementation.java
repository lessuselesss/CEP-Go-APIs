package circular.enterprise.apis;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import okhttp3.*;
import org.bouncycastle.asn1.x9.X9ECParameters;
import org.bouncycastle.crypto.ec.CustomNamedCurves;
import org.bouncycastle.crypto.params.ECDomainParameters;
import org.bouncycastle.crypto.params.ECPrivateKeyParameters;
import org.bouncycastle.crypto.signers.ECDSASigner;
import org.bouncycastle.crypto.signers.HMacDSAKCalculator;
import org.bouncycastle.jce.provider.BouncyCastleProvider;
import org.bouncycastle.math.ec.FixedPointCombMultiplier;

import java.math.BigInteger;
import java.security.MessageDigest;
import java.security.Security;
import java.time.Instant;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.TimeUnit;

/**
 * Circular Account Class
 */
public class CEPAccount {
    private String address;
    private String publicKey;
    private Map<String, Object> info;
    private final String codeVersion;
    private String lastError;
    private String nagUrl;
    private String networkNode;
    private String blockchain;
    private String latestTxID;
    private long nonce;
    private Object[] data;
    private int intervalSec;
    private static final OkHttpClient client;
    private static final ObjectMapper objectMapper;
    private static final MediaType JSON = MediaType.parse("application/json; charset=utf-8");

    static {
        client = new OkHttpClient.Builder()
                .connectTimeout(30, TimeUnit.SECONDS)
                .readTimeout(30, TimeUnit.SECONDS)
                .writeTimeout(30, TimeUnit.SECONDS)
                .build();
        objectMapper = new ObjectMapper();
        Security.addProvider(new BouncyCastleProvider());
    }

    public CEPAccount() {
        this.address = null;
        this.publicKey = null;
        this.info = null;
        this.codeVersion = Helper.LIB_VERSION;
        this.lastError = "";
        this.nagUrl = Helper.DEFAULT_NAG;
        this.networkNode = "";
        this.blockchain = Helper.DEFAULT_CHAIN;
        this.latestTxID = "";
        this.nonce = 0;
        this.data = new Object[0];
        this.intervalSec = 2;
    }

    /**
     * Open an account by retrieving all the account info
     * @param address Account address
     * @return True if successful, False otherwise
     */
    public boolean open(String address) {
        if (address == null || address.isEmpty()) {
            this.lastError = "Invalid address";
            return false;
        }
        this.address = address;
        return true;
    }

    /**
     * Update the account data and Nonce field
     * @return True if successful, False otherwise
     */
    public boolean updateAccount() {
        if (this.address == null) {
            this.lastError = "Account not open";
            return false;
        }

        try {
            Map<String, String> data = new HashMap<>();
            data.put("Blockchain", Helper.hexFix(this.blockchain));
            data.put("Address", Helper.hexFix(this.address));
            data.put("Version", this.codeVersion);

            String jsonBody = objectMapper.writeValueAsString(data);
            RequestBody body = RequestBody.create(jsonBody, JSON);
            Request request = new Request.Builder()
                    .url(this.nagUrl + "Circular_GetWalletNonce_")
                    .post(body)
                    .build();

            try (Response response = client.newCall(request).execute()) {
                if (!response.isSuccessful()) throw new Exception("Network error: " + response.code());
                
                String responseBody = response.body().string();
                Map<String, Object> responseData = objectMapper.readValue(responseBody, Map.class);
                
                if (responseData.get("Result").equals(200) && 
                    responseData.containsKey("Response") && 
                    ((Map)responseData.get("Response")).containsKey("Nonce")) {
                    this.nonce = ((Number)((Map)responseData.get("Response")).get("Nonce")).longValue() + 1;
                    return true;
                } else {
                    this.lastError = "Invalid response format or missing Nonce field";
                    return false;
                }
            }
        } catch (Exception e) {
            this.lastError = "Error: " + e.getMessage();
            return false;
        }
    }

    /**
     * Set the blockchain network
     * @param network Network name (e.g., 'devnet', 'testnet', 'mainnet')
     * @return URL of the network
     * @throws Exception if network URL cannot be fetched
     */
    public String setNetwork(String network) throws Exception {
        String nagUrl = Helper.NETWORK_URL + network;
        System.out.println("Fetching network info from: " + nagUrl);
        Request request = new Request.Builder()
                .url(nagUrl)
                .get()
                .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) throw new Exception("Network error: " + response.code());
            
            String responseBody = response.body().string();
            System.out.println("Network response: " + responseBody);
            Map<String, Object> data = objectMapper.readValue(responseBody, new TypeReference<Map<String, Object>>() {});
            this.nagUrl = data.get("url");
            return (String) data.get("url");
        }
    }

    /**
     * Set the blockchain address
     * @param blockchain Blockchain address
     */
    public void setBlockchain(String blockchain) {
        this.blockchain = blockchain;
    }

    /**
     * Close the account
     */
    public void close() {
        this.address = null;
        this.publicKey = null;
        this.info = null;
        this.lastError = "";
        this.nagUrl = null;
        this.networkNode = null;
        this.blockchain = null;
        this.latestTxID = null;
        this.data = null;
        this.nonce = 0;
        this.intervalSec = 0;
    }

    /**
     * Sign data using the account's private key
     * @param message Message to sign
     * @param privateKeyHex Private key in hex format
     * @return Signature in hex format
     * @throws Exception if signing fails
     */
    private String signData(String message, String privateKeyHex) throws Exception {
        if (this.address == null) {
            throw new Exception("Account is not open");
        }

        // Get curve parameters
        X9ECParameters curve = CustomNamedCurves.getByName("secp256k1");
        ECDomainParameters domain = new ECDomainParameters(curve.getCurve(), curve.getG(), curve.getN(), curve.getH());

        // Create signer with RFC 6979 deterministic k
        ECDSASigner signer = new ECDSASigner(new HMacDSAKCalculator(new org.bouncycastle.crypto.digests.SHA256Digest()));
        
        // Set up private key - ensure it's properly formatted
        String cleanPrivateKey = Helper.hexFix(privateKeyHex);
        if (cleanPrivateKey.length() != 64) {
            throw new Exception("Invalid private key length. Expected 64 characters (32 bytes)");
        }
        BigInteger privateKey = new BigInteger(cleanPrivateKey, 16);
        ECPrivateKeyParameters privateKeyParams = new ECPrivateKeyParameters(privateKey, domain);
        signer.init(true, privateKeyParams);

        // Hash the message - ensure UTF-8 encoding
        MessageDigest digest = MessageDigest.getInstance("SHA-256");
        byte[] messageHash = digest.digest(message.getBytes("UTF-8"));

        // Ensure the hash is treated as a positive number
        BigInteger messageHashBigInt = new BigInteger(1, messageHash);
        
        // Sign the hash
        BigInteger[] signature = signer.generateSignature(messageHash);
        
        // Normalize S value to be in lower half of curve order
        BigInteger halfN = domain.getN().shiftRight(1);
        if (signature[1].compareTo(halfN) > 0) {
            signature[1] = domain.getN().subtract(signature[1]);
        }
        
        // Convert to DER format
        byte[] derSignature = toDERFormat(signature[0], signature[1]);
        
        // Convert to hex
        return bytesToHex(derSignature);
    }

    /**
     * Get transaction by ID
     * @param txId Transaction ID
     * @param start Start block
     * @param end End block
     * @return Transaction data
     * @throws Exception if request fails
     */
    private Map<String, Object> getTransactionById(String txId, long start, long end) throws Exception {
        Map<String, String> data = new HashMap<>();
        data.put("Blockchain", Helper.hexFix(this.blockchain));
        data.put("ID", Helper.hexFix(txId));
        data.put("Start", String.valueOf(start));
        data.put("End", String.valueOf(end));
        data.put("Version", this.codeVersion);

        String url = this.nagUrl + "Circular_GetTransactionbyID_" + this.networkNode;
        String jsonBody = objectMapper.writeValueAsString(data);
        RequestBody body = RequestBody.create(jsonBody, JSON);
        
        Request request = new Request.Builder()
                .url(url)
                .post(body)
                .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) throw new Exception("Network error: " + response.code());
            return objectMapper.readValue(response.body().string(), Map.class);
        }
    }

        /**
     * Get transaction by ID and BlockNumner
     * @param txId Transaction ID
     * @param start Start block
     * @param end End block
     * @return Transaction data
     * @throws Exception if request fails
     */
    public Map<String, Object> getTransaction(String BlockID, String txId) throws Exception {
        Map<String, String> data = new HashMap<>();
        data.put("Blockchain", Helper.hexFix(this.blockchain));
        data.put("ID", Helper.hexFix(txId));
        data.put("Start", BlockID);
        data.put("End", BlockID);
        data.put("Version", this.codeVersion);

        String url = this.nagUrl + "Circular_GetTransactionbyID_" + this.networkNode;
        String jsonBody = objectMapper.writeValueAsString(data);
        RequestBody body = RequestBody.create(jsonBody, JSON);
        
        Request request = new Request.Builder()
                .url(url)
                .post(body)
                .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) throw new Exception("Network error: " + response.code());
            return objectMapper.readValue(response.body().string(), Map.class);
        }
    }

    /**
     * Get transaction outcome with polling
     * @param txId Transaction ID
     * @param timeoutSec Timeout in seconds
     * @param intervalSec Polling interval in seconds
     * @return Transaction outcome
     * @throws Exception if timeout or other error occurs
     */
    public Map<String, Object> getTransactionOutcome(String txId, int timeoutSec, int intervalSec) throws Exception {
        Instant startTime = Instant.now();
        
        while (true) {
            if (Instant.now().getEpochSecond() - startTime.getEpochSecond() > timeoutSec) {
                throw new Exception("Timeout exceeded");
            }

            Map<String, Object> data = getTransactionById(txId, 0, 10);
            
            if ((Integer)data.get("Result") == 200 && 
                !data.get("Response").equals("Transaction Not Found") &&
                !((Map)data.get("Response")).get("Status").equals("Pending")) {
                return data;
            }
            
            Thread.sleep(intervalSec * 1000L);
        }
    }

    /**
     * Submit a certificate
     * @param pdata Certificate data
     * @param privateKeyHex Private key in hex format
     * @throws Exception if submission fails
     */
    public void submitCertificate(String pdata, String privateKeyHex) throws Exception {
        if (this.address == null) {
            throw new Exception("Account is not open");
        }

        // Create the initial payload object and convert to hex
        Map<String, String> payloadObject = new HashMap<>();
        payloadObject.put("Action", "CP_CERTIFICATE");
        payloadObject.put("Data", Helper.stringToHex(pdata));
        
        String jsonStr = objectMapper.writeValueAsString(payloadObject);
        String payload = Helper.stringToHex(jsonStr);
        
        // Get current timestamp in the correct format
        String timestamp = Helper.getFormattedTimestamp();
        
        // Create the string for hashing (exactly as in PHP)
        String strToHash = Helper.hexFix(this.blockchain) + 
                          Helper.hexFix(this.address) + 
                          Helper.hexFix(this.address) + 
                          payload + 
                          String.valueOf(this.nonce) + 
                          timestamp;
        
        // Generate the ID using SHA-256
        MessageDigest digest = MessageDigest.getInstance("SHA-256");
        String id = bytesToHex(digest.digest(strToHash.getBytes("UTF-8")));
        
        // Sign the ID (not the payload)
        String signature = signData(id, privateKeyHex);
        
        // Create the complete transaction data
        Map<String, String> transactionData = new HashMap<>();
        transactionData.put("ID", id);
        transactionData.put("From", Helper.hexFix(this.address));
        transactionData.put("To", Helper.hexFix(this.address));
        transactionData.put("Timestamp", timestamp);
        transactionData.put("Payload", payload);
        transactionData.put("Nonce", String.valueOf(this.nonce));
        transactionData.put("Signature", signature);
        transactionData.put("Blockchain", Helper.hexFix(this.blockchain));
        transactionData.put("Type", "C_TYPE_CERTIFICATE");
        transactionData.put("Version", this.codeVersion);

        // Submit the certificate
        String jsonBody = objectMapper.writeValueAsString(transactionData);
        System.out.println("Submitting transaction: " + jsonBody);
        RequestBody body = RequestBody.create(jsonBody, JSON);
        Request request = new Request.Builder()
                .url(this.nagUrl + "Circular_AddTransaction_" + this.networkNode)
                .post(body)
                .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) throw new Exception("Network error: " + response.code());
            
            String responseBody = response.body().string();
            System.out.println("Response: " + responseBody);
            Map<String, Object> responseData = objectMapper.readValue(responseBody, Map.class);
            
            int resultCode = ((Number) responseData.get("Result")).intValue();
            if (resultCode == 200) {
                // Save our generated transaction ID
                this.latestTxID = id;
                System.out.println("Transaction ID: " + this.latestTxID);
                // Increment nonce for next transaction
                this.nonce++;
            } else {
                throw new Exception("Certificate submission failed: " + responseData.get("Response"));
            }
        }
    }

    // Helper methods for signature encoding
    private static byte[] toDERFormat(BigInteger r, BigInteger s) {
        // DER format: 0x30 [total-length] 0x02 [r-length] [r] 0x02 [s-length] [s]
        byte[] rBytes = r.toByteArray();
        byte[] sBytes = s.toByteArray();

        // Ensure positive numbers by prepending 0x00 if needed
        if (rBytes[0] < 0) {
            byte[] temp = new byte[rBytes.length + 1];
            temp[0] = 0;
            System.arraycopy(rBytes, 0, temp, 1, rBytes.length);
            rBytes = temp;
        }
        if (sBytes[0] < 0) {
            byte[] temp = new byte[sBytes.length + 1];
            temp[0] = 0;
            System.arraycopy(sBytes, 0, temp, 1, sBytes.length);
            sBytes = temp;
        }

        int totalLength = 2 + rBytes.length + 2 + sBytes.length;
        byte[] derSignature = new byte[2 + totalLength];
        
        // Sequence tag
        derSignature[0] = 0x30;
        // Total length
        derSignature[1] = (byte) totalLength;
        // Integer tag for r
        derSignature[2] = 0x02;
        // r length
        derSignature[3] = (byte) rBytes.length;
        // r value
        System.arraycopy(rBytes, 0, derSignature, 4, rBytes.length);
        // Integer tag for s
        derSignature[4 + rBytes.length] = 0x02;
        // s length
        derSignature[5 + rBytes.length] = (byte) sBytes.length;
        // s value
        System.arraycopy(sBytes, 0, derSignature, 6 + rBytes.length, sBytes.length);

        return derSignature;
    }

    private static String bytesToHex(byte[] bytes) {
        StringBuilder result = new StringBuilder();
        for (byte b : bytes) {
            result.append(String.format("%02x", b));
        }
        return result.toString();
    }

    // Getters and setters
    public String getLastError() {
        return lastError;
    }

    public void setNagUrl(String nagUrl) {
        this.nagUrl = nagUrl;
    }

    public void setNetworkNode(String networkNode) {
        this.networkNode = networkNode;
    }

    public String getLatestTxID() {
        return latestTxID;
    }

    public void setLatestTxID(String latestTxID) {
        this.latestTxID = latestTxID;
    }

    public long getNonce() {
        return nonce;
    }

    public void setNonce(long nonce) {
        this.nonce = nonce;
    }

    public int getIntervalSec() {
        return intervalSec;
    }

    public void setIntervalSec(int intervalSec) {
        this.intervalSec = intervalSec;
    }
} 