package certificate

// /*
//  *    insert application data into the certificate
//  *    data : data content
//  */
//     SetData(data)
//     {
//          this.data = stringToHex(data);
//     }

// /*
//  *    extracts application data from the certificate
//  */
//     GetData()
//     {
//          return hexToString(this.data);
//     }

// /*
//  *    returns the certificate in JSON format
//  */
//     GetJSONCertificate(){
//         let certificate = {
//             "data": this.data,
//             "previousTxID": this.previousTxID,
//             "previousBlock": this.previousBlock,
//             "version": this.codeVersion
//         };
//         return JSON.stringify(certificate);
//     }

// /*
//  *    extracts certificate size
//  *    FIX: This function now correctly calculates the byte size of the certificate,
//  *    which is crucial when it contains multi-byte Unicode characters.
//  */
//     GetCertificateSize() {
//         let certificate = {
//             "data": this.data,
//             "previousTxID": this.previousTxID,
//             "previousBlock": this.previousBlock,
//             "version": this.codeVersion
//         };
//         const jsonString = JSON.stringify(certificate);
//         // Use Buffer.byteLength with 'utf8' to get the actual byte count,
//         // as string.length would give an incorrect result for multi-byte characters.
//         return Buffer.byteLength(jsonString, 'utf8');
//     }

// }
