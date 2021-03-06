/*
 * Copyright (C) 2017-2018 GIG Technology NV and Contributors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package devcert

//go:generate openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 1460 -nodes -subj "/C=BE/ST=Gent/L=Lochristi/O=0-stor/CN=dev.0stor.com"
//go:generate openssl ecparam -out jwt_key.pem -name secp384r1 -genkey -noout
//go:generate openssl ec -in jwt_key.pem -pubout -out jwt_pub.pem
