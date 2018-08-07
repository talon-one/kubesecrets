package kubesecrets

import (
	"fmt"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Secret struct {
	Name       string
	Type       string
	Data       map[string][]byte
	StringData map[string]string
	CreatedAt  time.Time
}

func secretFromKubeSecret(s *v1.Secret) (secret Secret, err error) {
	secret.Name = s.Name
	secret.Type = string(s.Type)
	secret.CreatedAt = s.CreationTimestamp.Time
	secret.Data = s.Data
	secret.StringData = make(map[string]string)

	for key, value := range s.Data {
		ignore := false
		for _, b := range value {
			if b < 32 || b > 127 {
				ignore = true
				break
			}
		}
		if !ignore {
			secret.StringData[key] = string(value)
		} else {
			secret.StringData[key] = "[Binary Data]"
		}
	}

	return secret, nil
}

func GetSecrets(namespace string, clientSet *kubernetes.Clientset, filter ...string) ([]Secret, error) {
	secretList, err := clientSet.CoreV1().Secrets(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var secrets []Secret

	var matchesFilter func(name string) bool

	if size := len(filter); size > 0 {
		filterCopy := make([]string, size)
		for i, f := range filter {
			filterCopy[i] = strings.ToLower(f)
		}
		matchesFilter = func(name string) bool {
			name = strings.ToLower(name)
			for _, f := range filterCopy {
				if f == name || strings.Contains(name, f) {
					return true
				}
			}
			return false
		}
	} else {
		matchesFilter = func(name string) bool {
			return true
		}
	}
	var secret Secret
	for _, i := range secretList.Items {
		if matchesFilter(i.Name) {
			secret, err = secretFromKubeSecret(&i)
			if err != nil {
				return nil, err
			}
			secrets = append(secrets, secret)
		}
	}
	return secrets, nil
}

func SetSecret(namespace string, clientSet *kubernetes.Clientset, secretName string, data []byte) (Secret, error) {
	name, keyName := getName(secretName)
	if name == "" || keyName == "" {
		return Secret{}, fmt.Errorf("invalid secret name: `%s'", secretName)
	}
	secrets := clientSet.CoreV1().Secrets(namespace)

	// check if we need to update
	secretList, err := secrets.List(metav1.ListOptions{})
	if err != nil {
		return Secret{}, err
	}

	var kubeSecret *v1.Secret
	for _, i := range secretList.Items {
		if strings.EqualFold(i.Name, name) {
			i.Data[keyName] = data
			kubeSecret, err = secrets.Update(&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Data: i.Data,
			})
			if err != nil {
				return Secret{}, err
			}
			return secretFromKubeSecret(kubeSecret)
		}
	}
	kubeSecret, err = secrets.Create(&v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			keyName: data,
		},
	})
	if err != nil {
		return Secret{}, err
	}
	return secretFromKubeSecret(kubeSecret)
}

func DeleteSecret(namespace string, clientSet *kubernetes.Clientset, secretName string) (Secret, error) {
	name, keyName := getName(secretName)
	if name == "" {
		return Secret{}, fmt.Errorf("invalid secret name: `%s'", secretName)
	}

	secrets := clientSet.CoreV1().Secrets(namespace)

	// check if we need to update
	secretList, err := secrets.List(metav1.ListOptions{})
	if err != nil {
		return Secret{}, err
	}

	var kubeSecret *v1.Secret
	for _, i := range secretList.Items {
		if strings.EqualFold(i.Name, name) {
			if keyName == "" {
				// delete the whole item
				if err = secrets.Delete(name, &metav1.DeleteOptions{}); err != nil {
					return Secret{}, err
				}
				return secretFromKubeSecret(&i)
			} else {
				// delete the key only
				if _, ok := i.Data[keyName]; ok {
					delete(i.Data, keyName)
					kubeSecret, err = secrets.Update(&v1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      name,
							Namespace: namespace,
						},
						Data: i.Data,
					})
					if err != nil {
						return Secret{}, err
					}
					return secretFromKubeSecret(kubeSecret)
				}
				return secretFromKubeSecret(&i)
			}
		}
	}
	return Secret{}, fmt.Errorf("no secret found for `%s'", name)
}

func getName(name string) (secretName string, keyName string) {
	name = strings.TrimSpace(name)
	if i := strings.IndexRune(name, '.'); i > -1 {
		return name[:i], name[i+1:]
	}
	return name, ""
}
